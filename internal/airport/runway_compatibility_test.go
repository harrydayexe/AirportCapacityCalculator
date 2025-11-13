package airport

import (
	"strings"
	"testing"
)

func TestRunwayCompatibility_Validate_ValidConfiguration(t *testing.T) {
	// Test a valid simple crossing runway configuration
	runwayIDs := []string{"09L", "09R", "18"}
	compat := NewRunwayCompatibility(map[string][]string{
		"09L": {"09R"},
		"09R": {"09L"},
		"18":  {},
	})

	err := compat.Validate(runwayIDs)
	if err != nil {
		t.Errorf("Expected no error for valid configuration, got: %v", err)
	}
}

func TestRunwayCompatibility_Validate_NilCompatibility(t *testing.T) {
	// Nil compatibility should be valid (means all runways compatible)
	var compat *RunwayCompatibility
	runwayIDs := []string{"09L", "09R"}

	err := compat.Validate(runwayIDs)
	if err != nil {
		t.Errorf("Expected no error for nil compatibility, got: %v", err)
	}
}

func TestRunwayCompatibility_Validate_NilMap(t *testing.T) {
	// Compatibility with nil map should be valid
	compat := &RunwayCompatibility{CompatibleWith: nil}
	runwayIDs := []string{"09L", "09R"}

	err := compat.Validate(runwayIDs)
	if err != nil {
		t.Errorf("Expected no error for nil map, got: %v", err)
	}
}

func TestRunwayCompatibility_Validate_AsymmetricNoReverseList(t *testing.T) {
	// Test asymmetric relationship: A -> B but B has no compatibility list
	runwayIDs := []string{"09L", "09R"}
	compat := NewRunwayCompatibility(map[string][]string{
		"09L": {"09R"},
		// Missing "09R" entry
	})

	err := compat.Validate(runwayIDs)
	if err == nil {
		t.Error("Expected error for asymmetric compatibility (missing reverse list)")
	}
	if !strings.Contains(err.Error(), "asymmetric") {
		t.Errorf("Error should mention asymmetry, got: %v", err)
	}
}

func TestRunwayCompatibility_Validate_AsymmetricMissingReverseEntry(t *testing.T) {
	// Test asymmetric relationship: A -> B but B doesn't list A
	runwayIDs := []string{"09L", "09R", "18"}
	compat := NewRunwayCompatibility(map[string][]string{
		"09L": {"09R"},
		"09R": {},    // Has list but doesn't include 09L
		"18":  {},
	})

	err := compat.Validate(runwayIDs)
	if err == nil {
		t.Error("Expected error for asymmetric compatibility (missing reverse entry)")
	}
	if !strings.Contains(err.Error(), "asymmetric") {
		t.Errorf("Error should mention asymmetry, got: %v", err)
	}
}

func TestRunwayCompatibility_Validate_NonExistentRunwayInGraph(t *testing.T) {
	// Test compatibility graph referencing non-existent runway
	runwayIDs := []string{"09L", "09R"}
	compat := NewRunwayCompatibility(map[string][]string{
		"09L": {"09R"},
		"09R": {"09L"},
		"27":  {},    // Runway "27" doesn't exist in airport
	})

	err := compat.Validate(runwayIDs)
	if err == nil {
		t.Error("Expected error for non-existent runway in graph")
	}
	if !strings.Contains(err.Error(), "non-existent") {
		t.Errorf("Error should mention non-existent runway, got: %v", err)
	}
}

func TestRunwayCompatibility_Validate_NonExistentRunwayInCompatibleList(t *testing.T) {
	// Test compatible list referencing non-existent runway
	runwayIDs := []string{"09L", "09R"}
	compat := NewRunwayCompatibility(map[string][]string{
		"09L": {"09R", "27"},  // "27" doesn't exist
		"09R": {"09L"},
	})

	err := compat.Validate(runwayIDs)
	if err == nil {
		t.Error("Expected error for non-existent runway in compatible list")
	}
	if !strings.Contains(err.Error(), "non-existent") {
		t.Errorf("Error should mention non-existent runway, got: %v", err)
	}
}

func TestRunwayCompatibility_Validate_MissingRunwayFromGraph(t *testing.T) {
	// Test runway exists in airport but not in compatibility graph
	runwayIDs := []string{"09L", "09R", "18"}
	compat := NewRunwayCompatibility(map[string][]string{
		"09L": {"09R"},
		"09R": {"09L"},
		// Missing "18" entry
	})

	err := compat.Validate(runwayIDs)
	if err == nil {
		t.Error("Expected error for runway missing from graph")
	}
	if !strings.Contains(err.Error(), "not in the compatibility graph") {
		t.Errorf("Error should mention missing runway, got: %v", err)
	}
}

func TestRunwayCompatibility_Validate_SelfLoop(t *testing.T) {
	// Self-loops should be ignored (not cause errors)
	runwayIDs := []string{"09L", "09R"}
	compat := NewRunwayCompatibility(map[string][]string{
		"09L": {"09L", "09R"},  // Self-loop
		"09R": {"09L"},
	})

	err := compat.Validate(runwayIDs)
	if err != nil {
		t.Errorf("Expected no error for self-loop (should be ignored), got: %v", err)
	}
}

func TestRunwayCompatibility_IsCompatible_BothDirections(t *testing.T) {
	compat := NewRunwayCompatibility(map[string][]string{
		"09L": {"09R"},
		"09R": {"09L"},
		"18":  {},
	})

	// Test compatible pair (both directions)
	if !compat.IsCompatible("09L", "09R") {
		t.Error("09L should be compatible with 09R")
	}
	if !compat.IsCompatible("09R", "09L") {
		t.Error("09R should be compatible with 09L (symmetric)")
	}

	// Test incompatible pair
	if compat.IsCompatible("09L", "18") {
		t.Error("09L should not be compatible with 18")
	}
	if compat.IsCompatible("18", "09L") {
		t.Error("18 should not be compatible with 09L (symmetric)")
	}
}

func TestRunwayCompatibility_IsCompatible_SelfCompatibility(t *testing.T) {
	compat := NewRunwayCompatibility(map[string][]string{
		"09L": {"09R"},
		"09R": {"09L"},
	})

	// A runway is always compatible with itself
	if !compat.IsCompatible("09L", "09L") {
		t.Error("Runway should be compatible with itself")
	}
}

func TestRunwayCompatibility_IsCompatible_NilCompatibility(t *testing.T) {
	var compat *RunwayCompatibility

	// Nil compatibility means all runways compatible
	if !compat.IsCompatible("09L", "09R") {
		t.Error("With nil compatibility, all runways should be compatible")
	}
}

func TestRunwayCompatibility_IsCompatible_RunwayNotInGraph(t *testing.T) {
	compat := NewRunwayCompatibility(map[string][]string{
		"09L": {"09R"},
		"09R": {"09L"},
	})

	// Runway not in graph should be incompatible
	if compat.IsCompatible("27", "09L") {
		t.Error("Runway not in graph should be incompatible")
	}
}

func TestRunwayCompatibility_GetCompatibleRunways_Basic(t *testing.T) {
	compat := NewRunwayCompatibility(map[string][]string{
		"09L": {"09R"},
		"09R": {"09L"},
		"18":  {},
	})

	compatible := compat.GetCompatibleRunways("09L", []string{"09L", "09R", "18"})

	if len(compatible) != 1 {
		t.Errorf("Expected 1 compatible runway, got %d", len(compatible))
	}
	if len(compatible) > 0 && compatible[0] != "09R" {
		t.Errorf("Expected 09R, got %s", compatible[0])
	}
}

func TestRunwayCompatibility_GetCompatibleRunways_EmptyList(t *testing.T) {
	compat := NewRunwayCompatibility(map[string][]string{
		"09L": {"09R"},
		"09R": {"09L"},
		"18":  {},  // No compatible runways
	})

	compatible := compat.GetCompatibleRunways("18", []string{"09L", "09R", "18"})

	if len(compatible) != 0 {
		t.Errorf("Expected 0 compatible runways, got %d", len(compatible))
	}
}

func TestRunwayCompatibility_GetCompatibleRunways_NilCompatibility(t *testing.T) {
	var compat *RunwayCompatibility
	allRunways := []string{"09L", "09R", "18"}

	compatible := compat.GetCompatibleRunways("09L", allRunways)

	// Should return all other runways (not including self)
	if len(compatible) != 2 {
		t.Errorf("Expected 2 compatible runways, got %d", len(compatible))
	}

	// Verify self is not included
	for _, rwy := range compatible {
		if rwy == "09L" {
			t.Error("Compatible list should not include the runway itself")
		}
	}
}

func TestRunwayCompatibility_GetCompatibleRunways_RunwayNotInGraph(t *testing.T) {
	compat := NewRunwayCompatibility(map[string][]string{
		"09L": {"09R"},
		"09R": {"09L"},
	})

	compatible := compat.GetCompatibleRunways("27", []string{"09L", "09R"})

	// Runway not in graph should have no compatible runways
	if len(compatible) != 0 {
		t.Errorf("Expected 0 compatible runways for runway not in graph, got %d", len(compatible))
	}
}

func TestRunwayCompatibility_GetCompatibleRunways_ReturnsCopy(t *testing.T) {
	compat := NewRunwayCompatibility(map[string][]string{
		"09L": {"09R", "18"},
		"09R": {"09L", "18"},
		"18":  {"09L", "09R"},
	})

	compatible1 := compat.GetCompatibleRunways("09L", []string{"09L", "09R", "18"})
	compatible2 := compat.GetCompatibleRunways("09L", []string{"09L", "09R", "18"})

	// Modify one result
	compatible1[0] = "MODIFIED"

	// Other result should be unchanged
	if compatible2[0] == "MODIFIED" {
		t.Error("GetCompatibleRunways should return a copy, not a reference")
	}
}

func TestRunwayCompatibility_String_NilCompatibility(t *testing.T) {
	var compat *RunwayCompatibility
	str := compat.String()

	if !strings.Contains(str, "all runways compatible") {
		t.Errorf("String representation of nil compatibility should indicate all compatible, got: %s", str)
	}
}

func TestRunwayCompatibility_String_WithRunways(t *testing.T) {
	compat := NewRunwayCompatibility(map[string][]string{
		"09L": {"09R"},
		"09R": {"09L"},
		"18":  {},
	})

	str := compat.String()

	// Should contain all runway IDs
	if !strings.Contains(str, "09L") {
		t.Error("String should contain 09L")
	}
	if !strings.Contains(str, "09R") {
		t.Error("String should contain 09R")
	}
	if !strings.Contains(str, "18") {
		t.Error("String should contain 18")
	}
}
