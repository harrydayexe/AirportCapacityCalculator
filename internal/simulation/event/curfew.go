package event

import (
	"context"
	"time"
)

// CurfewStartEvent represents the beginning of a curfew period when operations must stop.
type CurfewStartEvent struct {
	timestamp time.Time
}

// NewCurfewStartEvent creates a new curfew start event.
func NewCurfewStartEvent(timestamp time.Time) *CurfewStartEvent {
	return &CurfewStartEvent{
		timestamp: timestamp,
	}
}

// Time returns when the curfew starts.
func (e *CurfewStartEvent) Time() time.Time {
	return e.timestamp
}

// Type returns the event type.
func (e *CurfewStartEvent) Type() EventType {
	return CurfewStartType
}

// Apply activates the curfew and triggers runway configuration recalculation.
// During curfew, no runways will be active.
func (e *CurfewStartEvent) Apply(ctx context.Context, world WorldState) error {
	// Update curfew status (for historical tracking)
	world.SetCurfewActive(true)

	// Notify RunwayManager and schedule configuration change event (will be empty config)
	return world.NotifyCurfewChange(true, e.timestamp)
}

// CurfewEndEvent represents the end of a curfew period when operations may resume.
type CurfewEndEvent struct {
	timestamp time.Time
}

// NewCurfewEndEvent creates a new curfew end event.
func NewCurfewEndEvent(timestamp time.Time) *CurfewEndEvent {
	return &CurfewEndEvent{
		timestamp: timestamp,
	}
}

// Time returns when the curfew ends.
func (e *CurfewEndEvent) Time() time.Time {
	return e.timestamp
}

// Type returns the event type.
func (e *CurfewEndEvent) Type() EventType {
	return CurfewEndType
}

// Apply deactivates the curfew and triggers runway configuration recalculation.
// Available runways will become active again.
func (e *CurfewEndEvent) Apply(ctx context.Context, world WorldState) error {
	// Update curfew status (for historical tracking)
	world.SetCurfewActive(false)

	// Notify RunwayManager and schedule configuration change event (will restore available runways)
	return world.NotifyCurfewChange(false, e.timestamp)
}
