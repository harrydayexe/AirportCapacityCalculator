package policy

import (
	"context"
	"fmt"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/event"
)

// RotationStrategy defines how runways are rotated to minimize noise impact.
type RotationStrategy int

const (
	// NoRotation means runways are used based on wind/efficiency only
	NoRotation RotationStrategy = iota

	// TimeBasedRotation rotates runway usage at fixed time intervals
	TimeBasedRotation

	// PreferentialRunway designates specific runways for noise abatement
	PreferentialRunway

	// NoiseOptimizedRotation rotates to minimize noise impact on communities
	NoiseOptimizedRotation
)

// String returns the string representation of the rotation strategy.
func (rs RotationStrategy) String() string {
	switch rs {
	case NoRotation:
		return "NoRotation"
	case TimeBasedRotation:
		return "TimeBasedRotation"
	case PreferentialRunway:
		return "PreferentialRunway"
	case NoiseOptimizedRotation:
		return "NoiseOptimizedRotation"
	default:
		return "Unknown"
	}
}

// RotationSchedule defines time-bounded windows when rotation policies apply.
// This allows rotation to be active only during specific hours or days (e.g., weekends).
// If nil, rotation applies for the entire simulation period.
type RotationSchedule struct {
	StartHour  int              // Hour of day when rotation starts (0-23)
	EndHour    int              // Hour of day when rotation ends (0-23)
	DaysOfWeek []time.Weekday   // Days when rotation applies (nil = all days)
}

// RotationPolicyConfiguration holds configuration for runway rotation policies.
type RotationPolicyConfiguration struct {
	efficiencyMap map[RotationStrategy]float32
}

// NewDefaultRotationPolicyConfiguration creates a new default rotation policy configuration
func NewDefaultRotationPolicyConfiguration() *RotationPolicyConfiguration {
	return &RotationPolicyConfiguration{
		efficiencyMap: map[RotationStrategy]float32{
			NoRotation:             1.0,
			TimeBasedRotation:      0.95,
			PreferentialRunway:     0.90,
			NoiseOptimizedRotation: 0.80,
		},
	}
}

// NewRotationPolicyConfiguration creates a new rotation policy configuration
func NewRotationPolicyConfiguration(efficiencyMap map[RotationStrategy]float32) *RotationPolicyConfiguration {
	return &RotationPolicyConfiguration{
		efficiencyMap: efficiencyMap,
	}
}

// RunwayRotationPolicy implements runway rotation strategies to distribute
// aircraft movements across different runways over time.
type RunwayRotationPolicy struct {
	strategy RotationStrategy             // The selected rotation strategy
	config   *RotationPolicyConfiguration // Configuration for efficiency adjustments
	schedule *RotationSchedule            // Optional time-bounded rotation schedule (nil = always active)
}

// NewRunwayRotationPolicy creates a new runway rotation policy.
func NewRunwayRotationPolicy(strategy RotationStrategy, config *RotationPolicyConfiguration) *RunwayRotationPolicy {
	return &RunwayRotationPolicy{
		strategy: strategy,
		config:   config,
	}
}

// NewDefaultRunwayRotationPolicy creates a new runway rotation policy with the default configuration
func NewDefaultRunwayRotationPolicy(strategy RotationStrategy) *RunwayRotationPolicy {
	return &RunwayRotationPolicy{
		strategy: strategy,
		config:   NewDefaultRotationPolicyConfiguration(),
		schedule: nil, // Always active
	}
}

// NewRunwayRotationPolicyWithSchedule creates a new runway rotation policy with a time-bounded schedule.
// The rotation will only apply during the specified time windows.
func NewRunwayRotationPolicyWithSchedule(strategy RotationStrategy, config *RotationPolicyConfiguration, schedule *RotationSchedule) *RunwayRotationPolicy {
	return &RunwayRotationPolicy{
		strategy: strategy,
		config:   config,
		schedule: schedule,
	}
}

// Name returns the policy name.
func (p *RunwayRotationPolicy) Name() string {
	return fmt.Sprintf("RunwayRotationPolicy(%s)", p.strategy.String())
}

// GenerateEvents generates rotation change events based on the policy configuration.
// Different strategies affect capacity by applying efficiency multipliers.
// Rotation strategies introduce overhead and constraints that reduce theoretical maximum capacity.
//
// If no schedule is provided, rotation is active for the entire simulation period.
// If a schedule is provided, rotation change events are generated to enable/disable
// the rotation multiplier during specified time windows.
func (p *RunwayRotationPolicy) GenerateEvents(ctx context.Context, world EventWorld) error {
	startTime := world.GetStartTime()
	endTime := world.GetEndTime()

	// Get efficiency multiplier based on rotation strategy
	var efficiencyMultiplier float32
	switch p.strategy {
	case NoRotation:
		// No modification needed - use runways as efficiently as possible
		// Efficiency: 100% (no penalty)
		efficiencyMultiplier = p.config.efficiencyMap[NoRotation]

	case TimeBasedRotation:
		// Time-based rotation introduces small overhead during transition periods
		// when runways are switched. This accounts for the time to communicate
		// the change to air traffic control and pilots.
		// Efficiency: 95% (5% capacity reduction)
		efficiencyMultiplier = p.config.efficiencyMap[TimeBasedRotation]

	case PreferentialRunway:
		// Preferential runway systems designate specific runways for noise abatement,
		// which may not always align with optimal wind conditions or traffic flow.
		// This introduces moderate efficiency penalties.
		// Efficiency: 90% (10% capacity reduction)
		efficiencyMultiplier = p.config.efficiencyMap[PreferentialRunway]

	case NoiseOptimizedRotation:
		// Noise-optimized rotation prioritizes minimizing community noise impact
		// over operational efficiency. This typically requires using less efficient
		// runway configurations and more restrictive flight paths.
		// Efficiency: 80% (20% capacity reduction)
		efficiencyMultiplier = p.config.efficiencyMap[NoiseOptimizedRotation]

	default:
		return fmt.Errorf("unknown rotation strategy: %v", p.strategy)
	}

	// If no schedule, rotation is always active
	if p.schedule == nil {
		// Schedule a rotation change event at the start of the simulation
		// This sets the efficiency multiplier for the entire simulation period
		world.ScheduleEvent(event.NewRotationChangeEvent(efficiencyMultiplier, startTime))
		return nil
	}

	// Generate time-bounded rotation events
	currentTime := startTime
	for currentTime.Before(endTime) {
		// Check if current day matches schedule
		if p.shouldApplyOnDay(currentTime.Weekday()) {
			// Calculate rotation start time for this day
			rotationStart := time.Date(
				currentTime.Year(), currentTime.Month(), currentTime.Day(),
				p.schedule.StartHour, 0, 0, 0, currentTime.Location(),
			)

			// Calculate rotation end time for this day
			rotationEnd := time.Date(
				currentTime.Year(), currentTime.Month(), currentTime.Day(),
				p.schedule.EndHour, 0, 0, 0, currentTime.Location(),
			)

			// Ensure times are within simulation bounds
			if rotationStart.After(startTime) && rotationStart.Before(endTime) {
				world.ScheduleEvent(event.NewRotationChangeEvent(efficiencyMultiplier, rotationStart))
			}

			if rotationEnd.After(startTime) && rotationEnd.Before(endTime) {
				// Return to 1.0 (no rotation penalty) when rotation window ends
				world.ScheduleEvent(event.NewRotationChangeEvent(1.0, rotationEnd))
			}
		}

		// Move to next day
		currentTime = currentTime.AddDate(0, 0, 1)
	}

	return nil
}

// shouldApplyOnDay checks if rotation should apply on the given weekday.
// Returns true if DaysOfWeek is nil (applies all days) or contains the given day.
func (p *RunwayRotationPolicy) shouldApplyOnDay(day time.Weekday) bool {
	if p.schedule.DaysOfWeek == nil {
		return true
	}

	for _, d := range p.schedule.DaysOfWeek {
		if d == day {
			return true
		}
	}
	return false
}
