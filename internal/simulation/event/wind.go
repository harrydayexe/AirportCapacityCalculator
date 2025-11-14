package event

import (
	"context"
	"time"
)

// WindChangeEvent represents a change in wind conditions during the simulation.
// When applied, it updates the world's wind state which triggers the RunwayManager
// to recalculate the active runway configuration based on new wind constraints.
type WindChangeEvent struct {
	speedKnots    float64   // Wind speed in knots
	directionTrue float64   // Wind direction in degrees true (0-360)
	timestamp     time.Time // When this wind change occurs
}

// NewWindChangeEvent creates a new wind change event.
//
// Parameters:
//   - speedKnots: Wind speed in knots (must be >= 0)
//   - directionTrue: Wind direction in degrees true (0-360, where 0/360 = north)
//   - timestamp: When this wind condition takes effect
//
// The event will call world.SetWind() which automatically:
//   - Updates stored wind values
//   - Notifies RunwayManager to recalculate configuration
//   - Filters runways by crosswind/tailwind limits
//   - Selects optimal runway directions (forward/reverse)
//   - Schedules ActiveRunwayConfigurationChangedEvent
func NewWindChangeEvent(speedKnots, directionTrue float64, timestamp time.Time) *WindChangeEvent {
	return &WindChangeEvent{
		speedKnots:    speedKnots,
		directionTrue: directionTrue,
		timestamp:     timestamp,
	}
}

// Time returns when the wind change occurs.
func (e *WindChangeEvent) Time() time.Time {
	return e.timestamp
}

// Type returns the event type.
func (e *WindChangeEvent) Type() EventType {
	return WindChangeType
}

// Apply updates the world's wind conditions and triggers runway reconfiguration.
// This will cause the RunwayManager to:
//   1. Filter runways by new wind constraints (crosswind/tailwind limits)
//   2. Determine optimal runway directions (prefer maximum headwind)
//   3. Select maximum-capacity configuration from usable runways
//   4. Generate ActiveRunwayConfigurationChangedEvent
func (e *WindChangeEvent) Apply(ctx context.Context, world WorldState) error {
	return world.SetWind(e.speedKnots, e.directionTrue)
}

// GetSpeed returns the wind speed in knots.
func (e *WindChangeEvent) GetSpeed() float64 {
	return e.speedKnots
}

// GetDirection returns the wind direction in degrees true.
func (e *WindChangeEvent) GetDirection() float64 {
	return e.directionTrue
}
