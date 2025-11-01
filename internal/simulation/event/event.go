// Package event defines the event system for state changes in the simulation.
package event

import (
	"context"
	"time"
)

// Event represents a state change that occurs at a specific time during the simulation.
// Events are processed chronologically to calculate capacity in discrete time windows.
type Event interface {
	// Time returns when this event occurs
	Time() time.Time

	// Type returns the type of event for logging and debugging
	Type() EventType

	// Apply processes the event and modifies the world state accordingly
	Apply(ctx context.Context, world WorldState) error
}

// EventType identifies the category of state change
type EventType int

const (
	// CurfewStartType indicates operations must cease
	CurfewStartType EventType = iota

	// CurfewEndType indicates operations may resume
	CurfewEndType

	// RunwayMaintenanceStartType indicates a runway becomes unavailable
	RunwayMaintenanceStartType

	// RunwayMaintenanceEndType indicates a runway becomes available
	RunwayMaintenanceEndType

	// RotationChangeType indicates rotation efficiency changes
	RotationChangeType

	// GateCapacityConstraintType indicates a gate capacity constraint is applied
	GateCapacityConstraintType
)

// String returns the string representation of the event type
func (et EventType) String() string {
	switch et {
	case CurfewStartType:
		return "CurfewStart"
	case CurfewEndType:
		return "CurfewEnd"
	case RunwayMaintenanceStartType:
		return "RunwayMaintenanceStart"
	case RunwayMaintenanceEndType:
		return "RunwayMaintenanceEnd"
	case RotationChangeType:
		return "RotationChange"
	case GateCapacityConstraintType:
		return "GateCapacityConstraint"
	default:
		return "Unknown"
	}
}

// WorldState defines the interface for accessing and modifying simulation state.
// This abstraction allows events to modify state without depending on the concrete type.
type WorldState interface {
	// SetCurfewActive sets whether curfew is currently active
	SetCurfewActive(active bool)

	// GetCurfewActive returns whether curfew is currently active
	GetCurfewActive() bool

	// SetRunwayAvailable marks a runway as available or unavailable
	SetRunwayAvailable(runwayID string, available bool) error

	// GetRunwayAvailable checks if a runway is currently available
	GetRunwayAvailable(runwayID string) (bool, error)

	// SetRotationMultiplier sets the current rotation efficiency multiplier
	SetRotationMultiplier(multiplier float32)

	// GetRotationMultiplier returns the current rotation efficiency multiplier
	GetRotationMultiplier() float32

	// SetGateCapacityConstraint sets the maximum movements per second allowed by gate capacity
	SetGateCapacityConstraint(maxMovementsPerSecond float32) error

	// GetGateCapacityConstraint returns the gate capacity constraint (0 means no constraint)
	GetGateCapacityConstraint() float32
}
