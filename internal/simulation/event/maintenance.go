package event

import (
	"context"
	"time"
)

// RunwayMaintenanceStartEvent represents a runway becoming unavailable for maintenance.
type RunwayMaintenanceStartEvent struct {
	runwayID  string
	timestamp time.Time
}

// NewRunwayMaintenanceStartEvent creates a new runway maintenance start event.
func NewRunwayMaintenanceStartEvent(runwayID string, timestamp time.Time) *RunwayMaintenanceStartEvent {
	return &RunwayMaintenanceStartEvent{
		runwayID:  runwayID,
		timestamp: timestamp,
	}
}

// Time returns when maintenance starts.
func (e *RunwayMaintenanceStartEvent) Time() time.Time {
	return e.timestamp
}

// Type returns the event type.
func (e *RunwayMaintenanceStartEvent) Type() EventType {
	return RunwayMaintenanceStartType
}

// RunwayID returns the ID of the runway undergoing maintenance.
func (e *RunwayMaintenanceStartEvent) RunwayID() string {
	return e.runwayID
}

// Apply marks the runway as unavailable.
func (e *RunwayMaintenanceStartEvent) Apply(ctx context.Context, world WorldState) error {
	return world.SetRunwayAvailable(e.runwayID, false)
}

// RunwayMaintenanceEndEvent represents a runway becoming available after maintenance.
type RunwayMaintenanceEndEvent struct {
	runwayID  string
	timestamp time.Time
}

// NewRunwayMaintenanceEndEvent creates a new runway maintenance end event.
func NewRunwayMaintenanceEndEvent(runwayID string, timestamp time.Time) *RunwayMaintenanceEndEvent {
	return &RunwayMaintenanceEndEvent{
		runwayID:  runwayID,
		timestamp: timestamp,
	}
}

// Time returns when maintenance ends.
func (e *RunwayMaintenanceEndEvent) Time() time.Time {
	return e.timestamp
}

// Type returns the event type.
func (e *RunwayMaintenanceEndEvent) Type() EventType {
	return RunwayMaintenanceEndType
}

// RunwayID returns the ID of the runway completing maintenance.
func (e *RunwayMaintenanceEndEvent) RunwayID() string {
	return e.runwayID
}

// Apply marks the runway as available.
func (e *RunwayMaintenanceEndEvent) Apply(ctx context.Context, world WorldState) error {
	return world.SetRunwayAvailable(e.runwayID, true)
}
