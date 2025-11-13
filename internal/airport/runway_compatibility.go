package airport

import (
	"fmt"
	"sort"
	"strings"
)

// RunwayCompatibility defines which runways can operate simultaneously.
// It uses a graph-based adjacency list representation where each runway
// has a list of other runways it can operate with.
//
// Example: At an airport with crossing runways 09L, 09R, and 18:
//   - 09L and 09R are parallel (compatible with each other)
//   - 18 crosses both 09L and 09R (incompatible with either)
//
// This would be represented as:
//
//	CompatibleWith: map[string][]string{
//	    "09L": {"09R"},
//	    "09R": {"09L"},
//	    "18":  {},  // Can operate alone but not with others
//	}
type RunwayCompatibility struct {
	// CompatibleWith maps each runway designation to a list of runways
	// it can operate with simultaneously.
	CompatibleWith map[string][]string
}

// NewRunwayCompatibility creates a new RunwayCompatibility instance.
func NewRunwayCompatibility(compatibleWith map[string][]string) *RunwayCompatibility {
	return &RunwayCompatibility{
		CompatibleWith: compatibleWith,
	}
}

// Validate checks that the compatibility graph is valid.
// It verifies:
//  1. Symmetry: If runway A is compatible with B, then B must be compatible with A
//  2. No invalid references: All referenced runways must exist in the airport's runway list
//  3. Self-loops are ignored (a runway is implicitly compatible with itself)
//
// Returns a descriptive error if validation fails, nil otherwise.
func (rc *RunwayCompatibility) Validate(runwayIDs []string) error {
	if rc == nil || rc.CompatibleWith == nil {
		return nil // nil compatibility is valid (means all runways compatible)
	}

	// Build a set of valid runway IDs for quick lookup
	validRunways := make(map[string]bool)
	for _, id := range runwayIDs {
		validRunways[id] = true
	}

	// Check each runway in the compatibility graph
	for runwayID, compatibleList := range rc.CompatibleWith {
		// Check that the runway itself exists
		if !validRunways[runwayID] {
			return fmt.Errorf("compatibility graph references non-existent runway: %s", runwayID)
		}

		// Check each runway in the compatible list
		for _, compatibleID := range compatibleList {
			// Ignore self-loops (runway compatible with itself)
			if compatibleID == runwayID {
				continue
			}

			// Check that referenced runway exists
			if !validRunways[compatibleID] {
				return fmt.Errorf("runway %s references non-existent compatible runway: %s",
					runwayID, compatibleID)
			}

			// Check symmetry: if A -> B, then B -> A must exist
			reverseList, exists := rc.CompatibleWith[compatibleID]
			if !exists {
				return fmt.Errorf("asymmetric compatibility: %s lists %s as compatible, but %s has no compatibility list",
					runwayID, compatibleID, compatibleID)
			}

			// Check if the reverse relationship exists
			reverseExists := false
			for _, reverseID := range reverseList {
				if reverseID == runwayID {
					reverseExists = true
					break
				}
			}

			if !reverseExists {
				return fmt.Errorf("asymmetric compatibility: %s lists %s as compatible, but %s does not list %s",
					runwayID, compatibleID, compatibleID, runwayID)
			}
		}
	}

	// Check that all runways have entries in the compatibility graph
	// (even if their compatible list is empty)
	for _, runwayID := range runwayIDs {
		if _, exists := rc.CompatibleWith[runwayID]; !exists {
			return fmt.Errorf("runway %s is not in the compatibility graph", runwayID)
		}
	}

	return nil
}

// IsCompatible checks if two runways can operate simultaneously.
// If compatibility is nil, returns true (all runways compatible).
// Self-compatibility always returns true.
func (rc *RunwayCompatibility) IsCompatible(runway1, runway2 string) bool {
	if rc == nil || rc.CompatibleWith == nil {
		return true // No compatibility defined means all compatible
	}

	if runway1 == runway2 {
		return true // A runway is always compatible with itself
	}

	compatibleList, exists := rc.CompatibleWith[runway1]
	if !exists {
		return false // If runway1 not in graph, incompatible
	}

	for _, compatibleID := range compatibleList {
		if compatibleID == runway2 {
			return true
		}
	}

	return false
}

// GetCompatibleRunways returns the list of runways compatible with the given runway.
// If compatibility is nil, returns all other runways in the provided list.
// The runway itself is not included in the result.
func (rc *RunwayCompatibility) GetCompatibleRunways(runwayID string, allRunways []string) []string {
	if rc == nil || rc.CompatibleWith == nil {
		// No compatibility defined, return all other runways
		result := make([]string, 0, len(allRunways)-1)
		for _, id := range allRunways {
			if id != runwayID {
				result = append(result, id)
			}
		}
		return result
	}

	compatibleList, exists := rc.CompatibleWith[runwayID]
	if !exists {
		return []string{} // Runway not in graph, no compatible runways
	}

	// Return a copy to prevent external modification
	result := make([]string, len(compatibleList))
	copy(result, compatibleList)
	return result
}

// String returns a human-readable representation of the compatibility graph.
func (rc *RunwayCompatibility) String() string {
	if rc == nil || rc.CompatibleWith == nil {
		return "RunwayCompatibility{all runways compatible}"
	}

	// Sort runway IDs for consistent output
	runways := make([]string, 0, len(rc.CompatibleWith))
	for runwayID := range rc.CompatibleWith {
		runways = append(runways, runwayID)
	}
	sort.Strings(runways)

	var builder strings.Builder
	builder.WriteString("RunwayCompatibility{\n")
	for _, runwayID := range runways {
		compatibleList := rc.CompatibleWith[runwayID]
		sort.Strings(compatibleList) // Sort for consistent output
		builder.WriteString(fmt.Sprintf("  %s: [%s]\n", runwayID, strings.Join(compatibleList, ", ")))
	}
	builder.WriteString("}")
	return builder.String()
}
