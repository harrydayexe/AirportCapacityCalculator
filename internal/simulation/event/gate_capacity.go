package event

import (
	"context"
	"time"
)

// GateCapacityConstraintEvent represents a gate capacity constraint being applied.
type GateCapacityConstraintEvent struct {
	maxMovementsPerSecond float32
	timestamp             time.Time
}

// NewGateCapacityConstraintEvent creates a new gate capacity constraint event.
func NewGateCapacityConstraintEvent(maxMovementsPerSecond float32, timestamp time.Time) *GateCapacityConstraintEvent {
	return &GateCapacityConstraintEvent{
		maxMovementsPerSecond: maxMovementsPerSecond,
		timestamp:             timestamp,
	}
}

// Time returns when the constraint is applied.
func (e *GateCapacityConstraintEvent) Time() time.Time {
	return e.timestamp
}

// Type returns the event type.
func (e *GateCapacityConstraintEvent) Type() EventType {
	return GateCapacityConstraintType
}

// MaxMovementsPerSecond returns the maximum movements per second allowed by gate capacity.
func (e *GateCapacityConstraintEvent) MaxMovementsPerSecond() float32 {
	return e.maxMovementsPerSecond
}

// Apply sets the gate capacity constraint in the world state.
func (e *GateCapacityConstraintEvent) Apply(ctx context.Context, world WorldState) error {
	return world.SetGateCapacityConstraint(e.maxMovementsPerSecond)
}
