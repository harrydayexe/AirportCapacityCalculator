package event

import (
	"context"
	"time"
)

// RotationChangeEvent represents a change in runway rotation strategy efficiency.
// Different rotation strategies apply different efficiency multipliers to capacity.
type RotationChangeEvent struct {
	multiplier float32
	timestamp  time.Time
}

// NewRotationChangeEvent creates a new rotation change event.
func NewRotationChangeEvent(multiplier float32, timestamp time.Time) *RotationChangeEvent {
	return &RotationChangeEvent{
		multiplier: multiplier,
		timestamp:  timestamp,
	}
}

// Time returns when the rotation change occurs.
func (e *RotationChangeEvent) Time() time.Time {
	return e.timestamp
}

// Type returns the event type.
func (e *RotationChangeEvent) Type() EventType {
	return RotationChangeType
}

// Multiplier returns the new efficiency multiplier.
func (e *RotationChangeEvent) Multiplier() float32 {
	return e.multiplier
}

// Apply updates the rotation efficiency multiplier.
func (e *RotationChangeEvent) Apply(ctx context.Context, world WorldState) error {
	world.SetRotationMultiplier(e.multiplier)
	return nil
}
