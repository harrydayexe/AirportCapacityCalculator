package simulation

import (
	"sync"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/airport"
	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/event"
)

// RunwayManager is responsible for managing runway availability and determining
// the active runway configuration. It is the single source of truth for which
// runways should be used for capacity calculations.
//
// Thread-Safety: All public methods are thread-safe and can be called concurrently.
// Uses sync.RWMutex for efficient concurrent reads with exclusive writes.
type RunwayManager struct {
	mu sync.RWMutex // Protects all fields below

	// availableRunways tracks which runways are physically available (not under maintenance)
	availableRunways map[string]bool

	// curfewActive indicates whether airport curfew is currently in effect
	curfewActive bool

	// allRunways contains the complete runway inventory for this airport
	allRunways []airport.Runway

	// currentConfiguration is the cached active runway configuration
	// Updated whenever availability or curfew status changes
	currentConfiguration map[string]*event.ActiveRunwayInfo
}

// NewRunwayManager creates a new thread-safe runway manager initialized with
// all runways available and no curfew active.
func NewRunwayManager(runways []airport.Runway) *RunwayManager {
	rm := &RunwayManager{
		availableRunways:     make(map[string]bool, len(runways)),
		curfewActive:         false,
		allRunways:           make([]airport.Runway, len(runways)),
		currentConfiguration: make(map[string]*event.ActiveRunwayInfo),
	}

	// Copy runways and initialize all as available
	copy(rm.allRunways, runways)
	for _, runway := range runways {
		rm.availableRunways[runway.RunwayDesignation] = true
	}

	// Calculate initial configuration
	rm.calculateActiveConfiguration()

	return rm
}

// OnRunwayAvailable notifies the manager that a runway has become available.
// This triggers recalculation of the active runway configuration.
//
// Thread-safe: Uses write lock.
func (rm *RunwayManager) OnRunwayAvailable(runwayID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.availableRunways[runwayID] = true
	rm.calculateActiveConfiguration()
}

// OnRunwayUnavailable notifies the manager that a runway has become unavailable.
// This triggers recalculation of the active runway configuration.
//
// Thread-safe: Uses write lock.
func (rm *RunwayManager) OnRunwayUnavailable(runwayID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.availableRunways[runwayID] = false
	rm.calculateActiveConfiguration()
}

// OnCurfewChanged notifies the manager that curfew status has changed.
// This triggers recalculation of the active runway configuration.
//
// Thread-safe: Uses write lock.
func (rm *RunwayManager) OnCurfewChanged(active bool) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.curfewActive = active
	rm.calculateActiveConfiguration()
}

// GetActiveConfiguration returns the current active runway configuration.
// Returns a deep copy to prevent external mutation of internal state.
//
// Thread-safe: Uses read lock.
func (rm *RunwayManager) GetActiveConfiguration() map[string]*event.ActiveRunwayInfo {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Return a deep copy to prevent external mutation
	config := make(map[string]*event.ActiveRunwayInfo, len(rm.currentConfiguration))
	for k, v := range rm.currentConfiguration {
		// Copy the struct (not just the pointer)
		infoCopy := *v
		config[k] = &infoCopy
	}

	return config
}

// calculateActiveConfiguration determines which runways should be active based on
// current availability and curfew status. This method updates currentConfiguration.
//
// Algorithm:
//  1. If curfew is active, no runways are active (return empty)
//  2. Get all available runways
//  3. TODO: Check for crossing runway conflicts (future enhancement)
//  4. TODO: Apply wind/direction logic (future enhancement)
//  5. Build active configuration with operation type and direction
//
// NOT thread-safe: Must be called while holding write lock (mu.Lock).
// This is a private method always called by lock-holding public methods.
func (rm *RunwayManager) calculateActiveConfiguration() {
	// Clear current configuration
	rm.currentConfiguration = make(map[string]*event.ActiveRunwayInfo)

	// If curfew is active, no runways are operational
	if rm.curfewActive {
		return
	}

	// Build active configuration from available runways
	for _, runway := range rm.allRunways {
		// Check if runway is available
		if available, exists := rm.availableRunways[runway.RunwayDesignation]; exists && available {
			// TODO: Check for crossing runway conflicts here
			// For now, all available runways are active

			// TODO: Determine operation type based on traffic patterns
			// For now, all runways handle mixed operations

			// TODO: Determine direction based on wind
			// For now, all runways use forward direction

			rm.currentConfiguration[runway.RunwayDesignation] = &event.ActiveRunwayInfo{
				RunwayDesignation: runway.RunwayDesignation,
				OperationType:     event.Mixed,    // Default: handle both takeoffs and landings
				Direction:         event.Forward,  // Default: primary direction
				Runway:            runway,
			}
		}
	}
}
