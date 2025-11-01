package simulation

import (
	"fmt"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/airport"
	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/event"
)

// World represents the complete state of the simulation at any point in time.
// It tracks runway availability, curfew status, and rotation efficiency.
type World struct {
	Airport     airport.Airport
	StartTime   time.Time
	EndTime     time.Time
	CurrentTime time.Time

	// Event queue for chronological processing
	Events *event.EventQueue

	// System state
	RunwayStates             map[string]*RunwayState
	CurfewActive             bool
	RotationMultiplier       float32
	GateCapacityConstraint   float32 // Max movements/second limited by gates (0 = no limit)

	// Statistics
	TotalCapacity float32
}

// RunwayState tracks a single runway's availability and operational parameters.
type RunwayState struct {
	Runway    airport.Runway
	Available bool
}

// NewWorld creates a new simulation world initialized with an airport.
func NewWorld(airport airport.Airport, startTime, endTime time.Time) *World {
	world := &World{
		Airport:            airport,
		StartTime:          startTime,
		EndTime:            endTime,
		CurrentTime:        startTime,
		Events:             event.NewEventQueue(),
		RunwayStates:       make(map[string]*RunwayState),
		CurfewActive:       false,
		RotationMultiplier: 1.0, // Default: no rotation penalty
		TotalCapacity:      0,
	}

	// Initialize runway states
	for _, runway := range airport.Runways {
		world.RunwayStates[runway.RunwayDesignation] = &RunwayState{
			Runway:    runway,
			Available: true, // All runways start available
		}
	}

	return world
}

// Implement WorldState interface for event processing

// SetCurfewActive sets whether curfew is currently active.
func (w *World) SetCurfewActive(active bool) {
	w.CurfewActive = active
}

// GetCurfewActive returns whether curfew is currently active.
func (w *World) GetCurfewActive() bool {
	return w.CurfewActive
}

// SetRunwayAvailable marks a runway as available or unavailable.
func (w *World) SetRunwayAvailable(runwayID string, available bool) error {
	state, exists := w.RunwayStates[runwayID]
	if !exists {
		return fmt.Errorf("runway %s not found", runwayID)
	}

	state.Available = available
	return nil
}

// GetRunwayAvailable checks if a runway is currently available.
func (w *World) GetRunwayAvailable(runwayID string) (bool, error) {
	state, exists := w.RunwayStates[runwayID]
	if !exists {
		return false, fmt.Errorf("runway %s not found", runwayID)
	}

	return state.Available, nil
}

// SetRotationMultiplier sets the current rotation efficiency multiplier.
func (w *World) SetRotationMultiplier(multiplier float32) {
	w.RotationMultiplier = multiplier
}

// GetRotationMultiplier returns the current rotation efficiency multiplier.
func (w *World) GetRotationMultiplier() float32 {
	return w.RotationMultiplier
}

// SetGateCapacityConstraint sets the maximum movements per second allowed by gate capacity.
func (w *World) SetGateCapacityConstraint(maxMovementsPerSecond float32) error {
	if maxMovementsPerSecond < 0 {
		return fmt.Errorf("gate capacity constraint cannot be negative: %f", maxMovementsPerSecond)
	}
	w.GateCapacityConstraint = maxMovementsPerSecond
	return nil
}

// GetGateCapacityConstraint returns the gate capacity constraint (0 means no constraint).
func (w *World) GetGateCapacityConstraint() float32 {
	return w.GateCapacityConstraint
}

// GetAvailableRunways returns a slice of currently available runways.
func (w *World) GetAvailableRunways() []airport.Runway {
	available := []airport.Runway{}

	for _, state := range w.RunwayStates {
		if state.Available {
			available = append(available, state.Runway)
		}
	}

	return available
}

// CountAvailableRunways returns the number of currently available runways.
func (w *World) CountAvailableRunways() int {
	count := 0
	for _, state := range w.RunwayStates {
		if state.Available {
			count++
		}
	}
	return count
}

// Implement EventWorld interface for policy interaction

// ScheduleEvent adds an event to the event queue.
func (w *World) ScheduleEvent(evt event.Event) {
	w.Events.Push(evt)
}

// GetEventQueue returns the event queue.
func (w *World) GetEventQueue() *event.EventQueue {
	return w.Events
}

// GetStartTime returns the simulation start time.
func (w *World) GetStartTime() time.Time {
	return w.StartTime
}

// GetEndTime returns the simulation end time.
func (w *World) GetEndTime() time.Time {
	return w.EndTime
}

// GetRunwayIDs returns a list of all runway IDs.
func (w *World) GetRunwayIDs() []string {
	ids := make([]string, 0, len(w.RunwayStates))
	for id := range w.RunwayStates {
		ids = append(ids, id)
	}
	return ids
}
