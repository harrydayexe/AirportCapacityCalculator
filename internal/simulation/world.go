// Package simulation provides the world state container and management for event-driven simulations.
// The World tracks all simulation state including runway availability, operational constraints,
// and capacity modifiers that affect the theoretical maximum throughput calculation.
package simulation

import (
	"fmt"
	"sync"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/airport"
	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/event"
)

// World represents the complete state of the simulation at any point in time.
// It tracks runway availability, curfew status, rotation efficiency, gate constraints,
// and taxi time overhead. The World is the central state container that events modify
// during the simulation to affect capacity calculations.
type World struct {
	// Airport configuration
	Airport airport.Airport // The airport being simulated

	// Time boundaries
	StartTime   time.Time // Simulation start time
	EndTime     time.Time // Simulation end time
	CurrentTime time.Time // Current simulation time (updated as events are processed)

	// Event processing
	Events *event.EventQueue // Priority queue of events ordered chronologically

	// Operational state
	RunwayStates map[string]*RunwayState // Per-runway availability and configuration (legacy, for historical tracking)
	CurfewActive bool                    // Whether airport curfew is currently in effect

	// Runway management (single source of truth for active runways)
	RunwayManager            *RunwayManager                          // Manages runway availability and active configuration
	activeConfigMu           sync.RWMutex                            // Protects ActiveRunwayConfiguration
	ActiveRunwayConfiguration map[string]*event.ActiveRunwayInfo     // Current active runway configuration

	// Capacity modifiers
	RotationMultiplier     float32       // Efficiency multiplier from runway rotation strategy (1.0 = no penalty)
	GateCapacityConstraint float32       // Max movements/second limited by gates (0 = no constraint)
	TaxiTimeOverhead       time.Duration // Total taxi time overhead per aircraft cycle (0 = no overhead)

	// Metrics
	TotalCapacity float32 // Accumulated total capacity (movements) calculated so far
}

// RunwayState tracks a single runway's operational status and configuration.
// Each runway in the airport has its own state that can be modified by events
// (e.g., maintenance makes runway unavailable).
type RunwayState struct {
	Runway    airport.Runway // The runway's configuration (designation, separation, etc.)
	Available bool           // Whether the runway is currently available for operations
}

// NewWorld creates a new simulation world initialized with an airport and time boundaries.
//
// The world is initialized with default values:
//   - All runways are available
//   - No curfew is active
//   - RotationMultiplier is 1.0 (no efficiency penalty)
//   - GateCapacityConstraint is 0 (no gate limitation)
//   - TaxiTimeOverhead is 0 (no taxi time impact)
//   - Empty event queue
//
// Policies will later modify these defaults by generating events that change the world state.
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

	// Initialize runway states - all runways start available
	for _, runway := range airport.Runways {
		world.RunwayStates[runway.RunwayDesignation] = &RunwayState{
			Runway:    runway,
			Available: true,
		}
	}

	// Initialize runway manager (single source of truth for active runways)
	world.RunwayManager = NewRunwayManager(airport.Runways, airport.RunwayCompatibility)

	// Set initial active runway configuration (all runways available)
	world.ActiveRunwayConfiguration = world.RunwayManager.GetActiveConfiguration()

	return world
}

// Implement WorldState interface for event processing.
// These methods are called by events when they are applied to the world state.

// SetCurfewActive sets whether airport curfew is currently in effect.
// Called by CurfewStartEvent (sets true) and CurfewEndEvent (sets false).
// When true, the engine will calculate zero capacity for the affected time window.
func (w *World) SetCurfewActive(active bool) {
	w.CurfewActive = active
}

// GetCurfewActive returns whether airport curfew is currently in effect.
func (w *World) GetCurfewActive() bool {
	return w.CurfewActive
}

// SetRunwayAvailable marks a runway as available or unavailable for operations.
// Called by RunwayMaintenanceStartEvent (sets false) and RunwayMaintenanceEndEvent (sets true).
// Unavailable runways are excluded from capacity calculations.
// Returns an error if the runway ID is not found in the airport configuration.
func (w *World) SetRunwayAvailable(runwayID string, available bool) error {
	state, exists := w.RunwayStates[runwayID]
	if !exists {
		return fmt.Errorf("runway %s not found", runwayID)
	}

	state.Available = available
	return nil
}

// GetRunwayAvailable checks if a runway is currently available for operations.
// Returns an error if the runway ID is not found in the airport configuration.
func (w *World) GetRunwayAvailable(runwayID string) (bool, error) {
	state, exists := w.RunwayStates[runwayID]
	if !exists {
		return false, fmt.Errorf("runway %s not found", runwayID)
	}

	return state.Available, nil
}

// SetRotationMultiplier sets the runway rotation efficiency multiplier.
// Called by RotationChangeEvent to apply efficiency penalties based on rotation strategy.
// Values < 1.0 represent efficiency loss (e.g., 0.95 = 5% penalty).
// Default is 1.0 (no penalty).
func (w *World) SetRotationMultiplier(multiplier float32) {
	w.RotationMultiplier = multiplier
}

// GetRotationMultiplier returns the current runway rotation efficiency multiplier.
func (w *World) GetRotationMultiplier() float32 {
	return w.RotationMultiplier
}

// SetGateCapacityConstraint sets the maximum movements per second allowed by gate capacity.
// Called by GateCapacityConstraintEvent during initialization.
// This constraint caps the sustained throughput when gates are more restrictive than runways.
// A value of 0 means no gate constraint is applied.
// Returns an error if the constraint is negative.
func (w *World) SetGateCapacityConstraint(maxMovementsPerSecond float32) error {
	if maxMovementsPerSecond < 0 {
		return fmt.Errorf("gate capacity constraint cannot be negative: %f", maxMovementsPerSecond)
	}
	w.GateCapacityConstraint = maxMovementsPerSecond
	return nil
}

// GetGateCapacityConstraint returns the gate capacity constraint in movements per second.
// A value of 0 means no constraint is applied.
func (w *World) GetGateCapacityConstraint() float32 {
	return w.GateCapacityConstraint
}

// SetTaxiTimeOverhead sets the total taxi time overhead per aircraft cycle.
// Called by TaxiTimeAdjustmentEvent during initialization.
// This overhead (taxi-in + taxi-out) extends the effective turnaround time, reducing
// the sustainable capacity when combined with gate constraints.
// A value of 0 means no taxi time impact.
// Returns an error if the overhead is negative.
func (w *World) SetTaxiTimeOverhead(overhead time.Duration) error {
	if overhead < 0 {
		return fmt.Errorf("taxi time overhead cannot be negative: %v", overhead)
	}
	w.TaxiTimeOverhead = overhead
	return nil
}

// GetTaxiTimeOverhead returns the taxi time overhead per aircraft cycle.
// A value of 0 means no taxi time overhead is applied.
func (w *World) GetTaxiTimeOverhead() time.Duration {
	return w.TaxiTimeOverhead
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

// SetActiveRunwayConfiguration sets the active runway configuration.
// This is the single source of truth for which runways the engine should use
// for capacity calculations. Stores a copy to prevent external mutation.
//
// Thread-safe: Uses write lock.
func (w *World) SetActiveRunwayConfiguration(config map[string]*event.ActiveRunwayInfo) error {
	w.activeConfigMu.Lock()
	defer w.activeConfigMu.Unlock()

	// Store a copy to prevent external mutation
	w.ActiveRunwayConfiguration = make(map[string]*event.ActiveRunwayInfo, len(config))
	for k, v := range config {
		// Deep copy the struct
		infoCopy := *v
		w.ActiveRunwayConfiguration[k] = &infoCopy
	}

	return nil
}

// GetActiveRunwayConfiguration returns the current active runway configuration.
// Returns a copy to prevent external mutation of internal state.
//
// Thread-safe: Uses read lock.
func (w *World) GetActiveRunwayConfiguration() map[string]*event.ActiveRunwayInfo {
	w.activeConfigMu.RLock()
	defer w.activeConfigMu.RUnlock()

	// Return a copy to prevent external mutation
	config := make(map[string]*event.ActiveRunwayInfo, len(w.ActiveRunwayConfiguration))
	for k, v := range w.ActiveRunwayConfiguration {
		// Deep copy the struct
		infoCopy := *v
		config[k] = &infoCopy
	}

	return config
}

// NotifyRunwayAvailabilityChange notifies the RunwayManager of a runway availability change
// and schedules an ActiveRunwayConfigurationChangedEvent with the new configuration.
// This ensures the active runway configuration is updated and the engine uses the correct runways.
func (w *World) NotifyRunwayAvailabilityChange(runwayID string, available bool, timestamp time.Time) error {
	// Notify the runway manager
	if available {
		w.RunwayManager.OnRunwayAvailable(runwayID)
	} else {
		w.RunwayManager.OnRunwayUnavailable(runwayID)
	}

	// Get the new active configuration from the manager
	newConfig := w.RunwayManager.GetActiveConfiguration()

	// Schedule an event to update the world's active configuration
	configEvent := event.NewActiveRunwayConfigurationChangedEvent(newConfig, timestamp)
	w.ScheduleEvent(configEvent)

	return nil
}

// NotifyCurfewChange notifies the RunwayManager of a curfew status change
// and schedules an ActiveRunwayConfigurationChangedEvent with the new configuration.
// During curfew, the configuration will be empty (no active runways).
func (w *World) NotifyCurfewChange(active bool, timestamp time.Time) error {
	// Notify the runway manager
	w.RunwayManager.OnCurfewChanged(active)

	// Get the new active configuration from the manager
	newConfig := w.RunwayManager.GetActiveConfiguration()

	// Schedule an event to update the world's active configuration
	configEvent := event.NewActiveRunwayConfigurationChangedEvent(newConfig, timestamp)
	w.ScheduleEvent(configEvent)

	return nil
}
