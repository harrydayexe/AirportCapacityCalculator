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

	// compatibility defines which runways can operate simultaneously (nil means all compatible)
	compatibility *airport.RunwayCompatibility

	// maximalCliques contains all maximal compatible runway sets (pre-computed for efficiency)
	maximalCliques [][]string

	// maximalCliquesComputed indicates whether maximal cliques have been computed
	maximalCliquesComputed bool
}

// NewRunwayManager creates a new thread-safe runway manager initialized with
// all runways available and no curfew active.
//
// Parameters:
//   - runways: The complete runway inventory for this airport
//   - compatibility: Optional runway compatibility graph (nil means all runways compatible)
func NewRunwayManager(runways []airport.Runway, compatibility *airport.RunwayCompatibility) *RunwayManager {
	rm := &RunwayManager{
		availableRunways:       make(map[string]bool, len(runways)),
		curfewActive:           false,
		allRunways:             make([]airport.Runway, len(runways)),
		currentConfiguration:   make(map[string]*event.ActiveRunwayInfo),
		compatibility:          compatibility,
		maximalCliques:         nil,
		maximalCliquesComputed: false,
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

// computeMaximalCliques finds all maximal compatible runway sets using Bron-Kerbosch algorithm.
// Maximal cliques represent the largest possible sets of runways that can operate together.
// This is computed lazily on first use and cached for subsequent calls.
//
// NOT thread-safe: Must be called while holding write lock.
func (rm *RunwayManager) computeMaximalCliques() {
	if rm.compatibility == nil {
		// No compatibility defined, all runways form one maximal clique
		allIDs := make([]string, 0, len(rm.allRunways))
		for _, runway := range rm.allRunways {
			allIDs = append(allIDs, runway.RunwayDesignation)
		}
		rm.maximalCliques = [][]string{allIDs}
		rm.maximalCliquesComputed = true
		return
	}

	// Build initial sets for Bron-Kerbosch
	// R = empty (current clique being built)
	// P = all vertices (candidates)
	// X = empty (already processed)
	R := []string{}
	P := make([]string, 0, len(rm.allRunways))
	X := []string{}

	for _, runway := range rm.allRunways {
		P = append(P, runway.RunwayDesignation)
	}

	result := make([][]string, 0)
	rm.bronKerbosch(R, P, X, &result)
	rm.maximalCliques = result
	rm.maximalCliquesComputed = true
}

// bronKerbosch implements the Bron-Kerbosch algorithm for finding all maximal cliques.
// This is a recursive backtracking algorithm.
//
// Parameters:
//   - R: Current clique being built
//   - P: Candidate vertices that could extend R
//   - X: Vertices already processed (excluded from further consideration)
//   - result: Accumulator for all maximal cliques found
//
// NOT thread-safe: Must be called while holding write lock.
func (rm *RunwayManager) bronKerbosch(R, P, X []string, result *[][]string) {
	// Base case: if P and X are both empty, R is a maximal clique
	if len(P) == 0 && len(X) == 0 {
		// Copy R to result (avoid reference issues)
		clique := make([]string, len(R))
		copy(clique, R)
		*result = append(*result, clique)
		return
	}

	// Iterate over a copy of P since we'll be modifying it
	PCopy := make([]string, len(P))
	copy(PCopy, P)

	for _, v := range PCopy {
		// Get neighbors of v (runways compatible with v)
		neighbors := rm.compatibility.GetCompatibleRunways(v, rm.getAllRunwayIDs())

		// R ∪ {v}
		newR := append([]string{}, R...)
		newR = append(newR, v)

		// P ∩ N(v)
		newP := intersection(P, neighbors)

		// X ∩ N(v)
		newX := intersection(X, neighbors)

		// Recursive call
		rm.bronKerbosch(newR, newP, newX, result)

		// Move v from P to X
		P = removeElement(P, v)
		X = append(X, v)
	}
}

// selectMaxCapacityConfig selects the compatible runway configuration with maximum capacity
// from the set of available runways.
//
// Algorithm:
//  1. Filter maximal cliques to only include those that are subsets of available runways
//  2. For each valid clique, calculate total capacity
//  3. Select the clique with highest capacity (prefer fewer runways on tie)
//
// Returns the runway IDs that should be active, or empty slice if no valid configuration.
//
// NOT thread-safe: Must be called while holding write lock.
func (rm *RunwayManager) selectMaxCapacityConfig(availableIDs []string) []string {
	if len(availableIDs) == 0 {
		return []string{}
	}

	// If no compatibility defined, return all available runways
	if rm.compatibility == nil {
		return availableIDs
	}

	// Ensure maximal cliques are computed
	if !rm.maximalCliquesComputed {
		rm.computeMaximalCliques()
	}

	// Find valid cliques (subsets of available runways)
	var bestConfig []string
	var bestCapacity float32 = 0

	for _, clique := range rm.maximalCliques {
		// Check if this clique is a subset of available runways
		if !isSubset(clique, availableIDs) {
			continue
		}

		// Calculate capacity for this configuration
		capacity := rm.calculateConfigCapacity(clique)

		// Select this config if:
		// 1. It has higher capacity, OR
		// 2. It has same capacity but fewer runways (simpler operations)
		if capacity > bestCapacity || (capacity == bestCapacity && len(clique) < len(bestConfig)) {
			bestCapacity = capacity
			bestConfig = clique
		}
	}

	return bestConfig
}

// calculateConfigCapacity calculates the total theoretical capacity for a runway configuration.
// Capacity is based on the sum of individual runway capacities (duration / separation time).
//
// For this calculation, we use a standard reference duration of 1 hour.
//
// NOT thread-safe: Must be called while holding write lock.
func (rm *RunwayManager) calculateConfigCapacity(runwayIDs []string) float32 {
	capacity := float32(0)
	const referenceDurationSeconds = 3600.0 // 1 hour

	for _, runwayID := range runwayIDs {
		runway, found := rm.findRunwayByID(runwayID)
		if !found {
			continue
		}

		separationSeconds := float32(runway.MinimumSeparation.Seconds())
		if separationSeconds > 0 {
			capacity += referenceDurationSeconds / separationSeconds
		}
	}

	return capacity
}

// getAvailableRunwayIDs returns a list of currently available runway IDs.
//
// NOT thread-safe: Must be called while holding read or write lock.
func (rm *RunwayManager) getAvailableRunwayIDs() []string {
	available := make([]string, 0, len(rm.availableRunways))
	for runwayID, isAvailable := range rm.availableRunways {
		if isAvailable {
			available = append(available, runwayID)
		}
	}
	return available
}

// getAllRunwayIDs returns a list of all runway IDs in the airport.
//
// NOT thread-safe: Must be called while holding read or write lock.
func (rm *RunwayManager) getAllRunwayIDs() []string {
	allIDs := make([]string, 0, len(rm.allRunways))
	for _, runway := range rm.allRunways {
		allIDs = append(allIDs, runway.RunwayDesignation)
	}
	return allIDs
}

// findRunwayByID finds a runway by its designation.
// Returns the runway and true if found, zero value and false otherwise.
//
// NOT thread-safe: Must be called while holding read or write lock.
func (rm *RunwayManager) findRunwayByID(runwayID string) (airport.Runway, bool) {
	for _, runway := range rm.allRunways {
		if runway.RunwayDesignation == runwayID {
			return runway, true
		}
	}
	return airport.Runway{}, false
}

// Helper functions for set operations

// intersection returns elements that appear in both slices.
func intersection(a, b []string) []string {
	set := make(map[string]bool)
	for _, item := range b {
		set[item] = true
	}

	result := make([]string, 0)
	for _, item := range a {
		if set[item] {
			result = append(result, item)
		}
	}
	return result
}

// isSubset checks if all elements of subset are in superset.
func isSubset(subset, superset []string) bool {
	superMap := make(map[string]bool)
	for _, item := range superset {
		superMap[item] = true
	}

	for _, item := range subset {
		if !superMap[item] {
			return false
		}
	}
	return true
}

// removeElement removes the first occurrence of an element from a slice.
func removeElement(slice []string, element string) []string {
	for i, item := range slice {
		if item == element {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// calculateActiveConfiguration determines which runways should be active based on
// current availability, curfew status, and runway compatibility constraints.
// This method updates currentConfiguration.
//
// Algorithm:
//  1. If curfew is active, no runways are active (return empty)
//  2. Get all available runways
//  3. Use compatibility graph to select maximum capacity configuration
//  4. Build active configuration with operation type and direction
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

	// Get available runway IDs
	availableIDs := rm.getAvailableRunwayIDs()

	// Select the optimal compatible configuration (maximum capacity)
	optimalConfig := rm.selectMaxCapacityConfig(availableIDs)

	// Build active configuration for the selected runways
	for _, runwayID := range optimalConfig {
		runway, found := rm.findRunwayByID(runwayID)
		if !found {
			continue
		}

		// TODO: Determine operation type based on traffic patterns
		// For now, all runways handle mixed operations

		// TODO: Determine direction based on wind
		// For now, all runways use forward direction

		rm.currentConfiguration[runwayID] = &event.ActiveRunwayInfo{
			RunwayDesignation: runwayID,
			OperationType:     event.Mixed,   // Default: handle both takeoffs and landings
			Direction:         event.Forward, // Default: primary direction
			Runway:            runway,
		}
	}
}
