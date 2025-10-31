package policy

import (
	"context"
	"errors"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/event"
)

// Common errors for curfew policy validation
var (
	// ErrInvalidCurfewTime indicates the curfew time range is invalid
	ErrInvalidCurfewTime = errors.New("curfew end time must be after start time")

	// ErrCurfewTooLong indicates the curfew duration exceeds reasonable limits
	ErrCurfewTooLong = errors.New("curfew duration exceeds maximum allowed duration")
)

const (
	// MaxCurfewDuration defines the maximum allowed curfew duration (30 days)
	// This prevents misconfiguration where extremely long curfews would make the simulation invalid
	MaxCurfewDuration = 30 * 24 * time.Hour
)

// EventWorld defines the interface for policies to interact with the simulation world.
// This interface is defined in the policy package to avoid circular dependencies.
type EventWorld interface {
	// Event queue management
	ScheduleEvent(event.Event)
	GetEventQueue() *event.EventQueue

	// Time boundaries
	GetStartTime() time.Time
	GetEndTime() time.Time

	// Runway information
	GetRunwayIDs() []string
}

// CurfewPolicy restricts airport operations during specified time ranges.
// It reduces the effective operating hours of the airport.
type CurfewPolicy struct {
	startTime time.Time // Start of curfew period
	endTime   time.Time // End of curfew period
}

// NewCurfewPolicy creates a new curfew policy with validation.
// Returns an error if the time range is invalid.
func NewCurfewPolicy(startTime, endTime time.Time) (*CurfewPolicy, error) {
	// Validate that end time is after start time
	if !endTime.After(startTime) {
		return nil, ErrInvalidCurfewTime
	}

	// Validate that the duration is reasonable
	duration := endTime.Sub(startTime)
	if duration > MaxCurfewDuration {
		return nil, ErrCurfewTooLong
	}

	return &CurfewPolicy{
		startTime: startTime,
		endTime:   endTime,
	}, nil
}

// Name returns the policy name.
func (p *CurfewPolicy) Name() string {
	return "CurfewPolicy"
}

// GenerateEvents generates curfew start and end events for every day in the simulation period.
// This implements the EventGeneratingPolicy interface for event-driven simulations.
func (p *CurfewPolicy) GenerateEvents(ctx context.Context, world EventWorld) error {
	startTime := world.GetStartTime()
	endTime := world.GetEndTime()

	// Extract hour and minute from the curfew times
	curfewStartHour, curfewStartMinute := p.startTime.Hour(), p.startTime.Minute()
	curfewEndHour, curfewEndMinute := p.endTime.Hour(), p.endTime.Minute()

	// Generate daily curfew events for the entire simulation period
	currentDate := startTime
	eventCount := 0

	for currentDate.Before(endTime) {
		// Create curfew start event for this day
		curfewStart := time.Date(
			currentDate.Year(), currentDate.Month(), currentDate.Day(),
			curfewStartHour, curfewStartMinute, 0, 0,
			currentDate.Location(),
		)

		// Only schedule if within simulation period
		if !curfewStart.Before(startTime) && !curfewStart.After(endTime) {
			world.ScheduleEvent(event.NewCurfewStartEvent(curfewStart))
			eventCount++
		}

		// Create curfew end event for this day (might be next day if overnight curfew)
		curfewEnd := time.Date(
			currentDate.Year(), currentDate.Month(), currentDate.Day(),
			curfewEndHour, curfewEndMinute, 0, 0,
			currentDate.Location(),
		)

		// Handle overnight curfews (end time is before start time)
		if curfewEndHour < curfewStartHour || (curfewEndHour == curfewStartHour && curfewEndMinute < curfewStartMinute) {
			curfewEnd = curfewEnd.AddDate(0, 0, 1)
		}

		// Only schedule if within simulation period (inclusive of end time)
		if !curfewEnd.Before(startTime) && !curfewEnd.After(endTime) {
			world.ScheduleEvent(event.NewCurfewEndEvent(curfewEnd))
			eventCount++
		}

		// Move to next day
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return nil
}
