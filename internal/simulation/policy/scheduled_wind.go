package policy

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/event"
)

// Common errors for scheduled wind policy validation
var (
	// ErrEmptyWindSchedule indicates no wind changes were provided
	ErrEmptyWindSchedule = errors.New("wind schedule cannot be empty")

	// ErrWindScheduleNotChronological indicates wind changes are not in time order
	ErrWindScheduleNotChronological = errors.New("wind schedule must be in chronological order")
)

// WindChange represents a discrete wind condition change at a specific time.
type WindChange struct {
	Timestamp     time.Time // When this wind condition takes effect
	SpeedKnots    float64   // Wind speed in knots
	DirectionTrue float64   // Wind direction in degrees true (0-360)
}

// ScheduledWindPolicy implements time-varying wind conditions based on an explicit schedule.
// Unlike the static WindPolicy, this generates WindChangeEvents at specified times
// to model realistic wind patterns such as diurnal cycles, frontal passages, or
// seasonal variations.
//
// Example use cases:
//   - Diurnal wind patterns (morning calm → afternoon sea breeze → evening shift)
//   - Weather fronts (abrupt wind direction changes)
//   - Seasonal prevailing winds
//   - Historical METAR data replay
//
// The schedule must:
//   - Be in chronological order
//   - Have valid wind parameters (speed >= 0, direction 0-360)
//   - Contain at least one wind change
type ScheduledWindPolicy struct {
	windSchedule []WindChange
}

// NewScheduledWindPolicy creates a new scheduled wind policy with validation.
//
// Validation rules:
//   - Schedule cannot be empty
//   - Wind changes must be in chronological order
//   - Wind speeds must be non-negative
//   - Wind directions are automatically normalized to 0-360 range
//
// Returns an error if validation fails.
func NewScheduledWindPolicy(windSchedule []WindChange) (*ScheduledWindPolicy, error) {
	if len(windSchedule) == 0 {
		return nil, ErrEmptyWindSchedule
	}

	// Validate and normalize wind changes
	for i, change := range windSchedule {
		// Validate speed
		if change.SpeedKnots < 0 {
			return nil, fmt.Errorf("wind change %d: %w", i, ErrInvalidWindSpeed)
		}

		// Normalize direction to 0-360 range
		normalizedDirection := math.Mod(change.DirectionTrue, 360)
		if normalizedDirection < 0 {
			normalizedDirection += 360
		}
		windSchedule[i].DirectionTrue = normalizedDirection

		// Check chronological order
		if i > 0 && !change.Timestamp.After(windSchedule[i-1].Timestamp) {
			return nil, ErrWindScheduleNotChronological
		}
	}

	return &ScheduledWindPolicy{
		windSchedule: windSchedule,
	}, nil
}

// Name returns the policy name.
func (p *ScheduledWindPolicy) Name() string {
	return "ScheduledWindPolicy"
}

// GenerateEvents creates WindChangeEvents for each scheduled wind change.
// Only generates events that fall within the simulation time period.
//
// The first wind change in the schedule sets the initial wind condition if it occurs
// at or before the simulation start time. Otherwise, the simulation starts with calm wind
// (0 knots) until the first scheduled change.
func (p *ScheduledWindPolicy) GenerateEvents(ctx context.Context, world EventWorld) error {
	startTime := world.GetStartTime()
	endTime := world.GetEndTime()

	eventCount := 0

	for _, change := range p.windSchedule {
		// Only schedule events within simulation period
		if change.Timestamp.Before(startTime) || change.Timestamp.After(endTime) {
			continue
		}

		// Create and schedule wind change event
		windEvent := event.NewWindChangeEvent(
			change.SpeedKnots,
			change.DirectionTrue,
			change.Timestamp,
		)

		world.ScheduleEvent(windEvent)
		eventCount++
	}

	return nil
}

// GetSchedule returns a copy of the wind schedule.
func (p *ScheduledWindPolicy) GetSchedule() []WindChange {
	schedule := make([]WindChange, len(p.windSchedule))
	copy(schedule, p.windSchedule)
	return schedule
}

// GetWindAt returns the wind conditions at a specific time based on the schedule.
// Returns the most recent wind change at or before the given time.
// If no wind change has occurred yet, returns calm wind (0 knots).
func (p *ScheduledWindPolicy) GetWindAt(timestamp time.Time) (speedKnots, directionTrue float64) {
	// Default to calm wind
	speedKnots = 0
	directionTrue = 0

	// Find the most recent wind change at or before the timestamp
	for _, change := range p.windSchedule {
		if change.Timestamp.After(timestamp) {
			break
		}
		speedKnots = change.SpeedKnots
		directionTrue = change.DirectionTrue
	}

	return speedKnots, directionTrue
}

// SortSchedule sorts the wind schedule chronologically in place.
// This is useful if you build a schedule programmatically and want to ensure
// chronological order before creating the policy.
func SortSchedule(schedule []WindChange) {
	sort.Slice(schedule, func(i, j int) bool {
		return schedule[i].Timestamp.Before(schedule[j].Timestamp)
	})
}
