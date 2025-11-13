package simulation

import (
	"sort"
	"testing"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/airport"
)

// Helper function to compare string slices without regard to order
func containsSameElements(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aCopy := make([]string, len(a))
	bCopy := make([]string, len(b))
	copy(aCopy, a)
	copy(bCopy, b)
	sort.Strings(aCopy)
	sort.Strings(bCopy)
	for i := range aCopy {
		if aCopy[i] != bCopy[i] {
			return false
		}
	}
	return true
}

// Test 1: Simple Crossing Runways
// Two perpendicular runways that cannot operate simultaneously
func TestRunwayManager_Compatibility_SimpleCrossing(t *testing.T) {
	runways := []airport.Runway{
		{RunwayDesignation: "09", TrueBearing: 90, MinimumSeparation: 90 * time.Second},
		{RunwayDesignation: "18", TrueBearing: 180, MinimumSeparation: 90 * time.Second},
	}

	// Runways cross each other, cannot operate together
	compat := airport.NewRunwayCompatibility(map[string][]string{
		"09": {},
		"18": {},
	})

	rm := NewRunwayManager(runways, compat)
	config := rm.GetActiveConfiguration()

	// Should select one runway (both have same capacity, prefer fewer = 1)
	if len(config) != 1 {
		t.Errorf("Expected 1 active runway, got %d", len(config))
	}

	// Either runway is acceptable
	if len(config) > 0 {
		activeRunway := ""
		for k := range config {
			activeRunway = k
		}
		if activeRunway != "09" && activeRunway != "18" {
			t.Errorf("Unexpected runway: %s", activeRunway)
		}
	}
}

// Test 2: Parallel Runways (All Compatible)
// Three parallel runways that can all operate together
func TestRunwayManager_Compatibility_ParallelRunways(t *testing.T) {
	runways := []airport.Runway{
		{RunwayDesignation: "09L", TrueBearing: 90, MinimumSeparation: 90 * time.Second},
		{RunwayDesignation: "09C", TrueBearing: 90, MinimumSeparation: 90 * time.Second},
		{RunwayDesignation: "09R", TrueBearing: 90, MinimumSeparation: 90 * time.Second},
	}

	// All parallel runways compatible
	compat := airport.NewRunwayCompatibility(map[string][]string{
		"09L": {"09C", "09R"},
		"09C": {"09L", "09R"},
		"09R": {"09L", "09C"},
	})

	rm := NewRunwayManager(runways, compat)
	config := rm.GetActiveConfiguration()

	// All three should be active
	if len(config) != 3 {
		t.Errorf("Expected 3 active runways, got %d", len(config))
	}

	expectedRunways := []string{"09L", "09C", "09R"}
	activeRunways := make([]string, 0, len(config))
	for k := range config {
		activeRunways = append(activeRunways, k)
	}

	if !containsSameElements(activeRunways, expectedRunways) {
		t.Errorf("Expected %v, got %v", expectedRunways, activeRunways)
	}
}

// Test 3: User's Specific Example
// Runway 1 allows Runway 2 OR Runway 3, but not both
func TestRunwayManager_Compatibility_UserExample(t *testing.T) {
	runways := []airport.Runway{
		{RunwayDesignation: "RWY1", TrueBearing: 90, MinimumSeparation: 90 * time.Second},
		{RunwayDesignation: "RWY2", TrueBearing: 180, MinimumSeparation: 90 * time.Second},
		{RunwayDesignation: "RWY3", TrueBearing: 270, MinimumSeparation: 90 * time.Second},
	}

	// RWY1 compatible with RWY2 and RWY3, but RWY2 and RWY3 not compatible with each other
	compat := airport.NewRunwayCompatibility(map[string][]string{
		"RWY1": {"RWY2", "RWY3"},
		"RWY2": {"RWY1"},
		"RWY3": {"RWY1"},
	})

	rm := NewRunwayManager(runways, compat)
	config := rm.GetActiveConfiguration()

	// Should select 2 runways (RWY1 + RWY2 or RWY1 + RWY3)
	if len(config) != 2 {
		t.Errorf("Expected 2 active runways, got %d", len(config))
	}

	// RWY1 must be in the configuration
	if _, exists := config["RWY1"]; !exists {
		t.Error("RWY1 should be in configuration")
	}

	// Either RWY2 or RWY3 should be present, but not both
	hasRWY2 := false
	hasRWY3 := false
	if _, exists := config["RWY2"]; exists {
		hasRWY2 = true
	}
	if _, exists := config["RWY3"]; exists {
		hasRWY3 = true
	}

	if hasRWY2 && hasRWY3 {
		t.Error("RWY2 and RWY3 should not both be active (incompatible)")
	}

	if !hasRWY2 && !hasRWY3 {
		t.Error("Either RWY2 or RWY3 should be active")
	}
}

// Test 4: Capacity-Based Selection (Higher Capacity Single Runway)
// One high-capacity runway vs two lower-capacity parallel runways
func TestRunwayManager_Compatibility_CapacityBasedSelection(t *testing.T) {
	runways := []airport.Runway{
		{RunwayDesignation: "09L", TrueBearing: 90, MinimumSeparation: 120 * time.Second}, // 30 mvmt/hr
		{RunwayDesignation: "09R", TrueBearing: 90, MinimumSeparation: 120 * time.Second}, // 30 mvmt/hr
		{RunwayDesignation: "18", TrueBearing: 180, MinimumSeparation: 48 * time.Second},  // 75 mvmt/hr
	}

	// 09L and 09R are parallel (compatible), 18 crosses both (incompatible)
	compat := airport.NewRunwayCompatibility(map[string][]string{
		"09L": {"09R"},
		"09R": {"09L"},
		"18":  {},
	})

	rm := NewRunwayManager(runways, compat)
	config := rm.GetActiveConfiguration()

	// Should select 18 alone (75 mvmt/hr) over 09L+09R (60 mvmt/hr combined)
	if len(config) != 1 {
		t.Errorf("Expected 1 active runway, got %d", len(config))
	}

	if _, exists := config["18"]; !exists {
		t.Error("Should select 18 (higher capacity) over 09L+09R")
	}
}

// Test 5: Tie-Breaking (Prefer Fewer Runways)
// Multiple configurations with same capacity - prefer simpler one
func TestRunwayManager_Compatibility_TieBreaking(t *testing.T) {
	runways := []airport.Runway{
		{RunwayDesignation: "09L", TrueBearing: 90, MinimumSeparation: 120 * time.Second},  // 30 mvmt/hr
		{RunwayDesignation: "09R", TrueBearing: 90, MinimumSeparation: 120 * time.Second},  // 30 mvmt/hr
		{RunwayDesignation: "18", TrueBearing: 180, MinimumSeparation: 60 * time.Second},   // 60 mvmt/hr
	}

	// 09L and 09R parallel, 18 crosses both
	compat := airport.NewRunwayCompatibility(map[string][]string{
		"09L": {"09R"},
		"09R": {"09L"},
		"18":  {},
	})

	rm := NewRunwayManager(runways, compat)
	config := rm.GetActiveConfiguration()

	// Both configs have 60 mvmt/hr: {09L, 09R} or {18}
	// Should prefer {18} (fewer runways)
	if len(config) != 1 {
		t.Errorf("Expected 1 active runway (tie-breaking: prefer fewer), got %d", len(config))
	}

	if _, exists := config["18"]; !exists {
		t.Error("Should prefer single runway 18 over two-runway configuration")
	}
}

// Test 6: Complex Airport (LAX-style)
// Four parallel runways with two groups
func TestRunwayManager_Compatibility_ComplexAirportLAX(t *testing.T) {
	runways := []airport.Runway{
		{RunwayDesignation: "24L", TrueBearing: 240, MinimumSeparation: 90 * time.Second},
		{RunwayDesignation: "24R", TrueBearing: 240, MinimumSeparation: 90 * time.Second},
		{RunwayDesignation: "25L", TrueBearing: 250, MinimumSeparation: 90 * time.Second},
		{RunwayDesignation: "25R", TrueBearing: 250, MinimumSeparation: 90 * time.Second},
	}

	// Two pairs of closely-spaced parallels
	// 24L/24R can operate together, 25L/25R can operate together
	// But 24x cannot operate with 25x (too close together)
	compat := airport.NewRunwayCompatibility(map[string][]string{
		"24L": {"24R"},
		"24R": {"24L"},
		"25L": {"25R"},
		"25R": {"25L"},
	})

	rm := NewRunwayManager(runways, compat)
	config := rm.GetActiveConfiguration()

	// Should select either {24L, 24R} or {25L, 25R} (both have same capacity)
	if len(config) != 2 {
		t.Errorf("Expected 2 active runways, got %d", len(config))
	}

	activeRunways := make([]string, 0, len(config))
	for k := range config {
		activeRunways = append(activeRunways, k)
	}

	valid24 := containsSameElements(activeRunways, []string{"24L", "24R"})
	valid25 := containsSameElements(activeRunways, []string{"25L", "25R"})

	if !valid24 && !valid25 {
		t.Errorf("Expected {24L, 24R} or {25L, 25R}, got %v", activeRunways)
	}
}

// Test 7: Backward Compatibility (Nil Compatibility)
// No compatibility defined means all runways can operate together
func TestRunwayManager_Compatibility_BackwardCompatibility(t *testing.T) {
	runways := []airport.Runway{
		{RunwayDesignation: "09", TrueBearing: 90, MinimumSeparation: 90 * time.Second},
		{RunwayDesignation: "18", TrueBearing: 180, MinimumSeparation: 90 * time.Second},
		{RunwayDesignation: "27", TrueBearing: 270, MinimumSeparation: 90 * time.Second},
	}

	// Nil compatibility = all runways compatible (old behavior)
	rm := NewRunwayManager(runways, nil)
	config := rm.GetActiveConfiguration()

	// All runways should be active
	if len(config) != 3 {
		t.Errorf("Expected 3 active runways with nil compatibility, got %d", len(config))
	}
}

// Test 8: Single Runway Airport
// Edge case: only one runway
func TestRunwayManager_Compatibility_SingleRunway(t *testing.T) {
	runways := []airport.Runway{
		{RunwayDesignation: "09", TrueBearing: 90, MinimumSeparation: 90 * time.Second},
	}

	compat := airport.NewRunwayCompatibility(map[string][]string{
		"09": {},
	})

	rm := NewRunwayManager(runways, compat)
	config := rm.GetActiveConfiguration()

	// Single runway should be active
	if len(config) != 1 {
		t.Errorf("Expected 1 active runway, got %d", len(config))
	}

	if _, exists := config["09"]; !exists {
		t.Error("Runway 09 should be active")
	}
}

// Test 9: Partial Runway Availability
// Some runways under maintenance
func TestRunwayManager_Compatibility_PartialAvailability(t *testing.T) {
	runways := []airport.Runway{
		{RunwayDesignation: "09L", TrueBearing: 90, MinimumSeparation: 90 * time.Second},
		{RunwayDesignation: "09R", TrueBearing: 90, MinimumSeparation: 90 * time.Second},
		{RunwayDesignation: "18", TrueBearing: 180, MinimumSeparation: 90 * time.Second},
	}

	// 09L and 09R parallel, 18 crosses
	compat := airport.NewRunwayCompatibility(map[string][]string{
		"09L": {"09R"},
		"09R": {"09L"},
		"18":  {},
	})

	rm := NewRunwayManager(runways, compat)

	// Mark 09R unavailable
	rm.OnRunwayUnavailable("09R")

	config := rm.GetActiveConfiguration()

	// Should select either 09L alone or 18 alone (same capacity)
	if len(config) != 1 {
		t.Errorf("Expected 1 active runway, got %d", len(config))
	}

	// Should be either 09L or 18, not 09R
	for k := range config {
		if k == "09R" {
			t.Error("09R should not be active (marked unavailable)")
		}
		if k != "09L" && k != "18" {
			t.Errorf("Unexpected runway: %s", k)
		}
	}
}

// Test 10: Changi-style (Complex Multi-Group)
// Six runways with complex compatibility
func TestRunwayManager_Compatibility_ChangiStyle(t *testing.T) {
	runways := []airport.Runway{
		{RunwayDesignation: "02L", TrueBearing: 20, MinimumSeparation: 90 * time.Second},
		{RunwayDesignation: "02C", TrueBearing: 20, MinimumSeparation: 90 * time.Second},
		{RunwayDesignation: "02R", TrueBearing: 20, MinimumSeparation: 90 * time.Second},
		{RunwayDesignation: "20L", TrueBearing: 200, MinimumSeparation: 90 * time.Second},
		{RunwayDesignation: "20C", TrueBearing: 200, MinimumSeparation: 90 * time.Second},
		{RunwayDesignation: "20R", TrueBearing: 200, MinimumSeparation: 90 * time.Second},
	}

	// Complex compatibility: can use 02s together OR 20s together, but not mix
	// Also some internal restrictions
	compat := airport.NewRunwayCompatibility(map[string][]string{
		"02L": {"02C"},
		"02C": {"02L"},
		"02R": {},
		"20L": {"20C"},
		"20C": {"20L"},
		"20R": {},
	})

	rm := NewRunwayManager(runways, compat)
	config := rm.GetActiveConfiguration()

	// Should select either {02L, 02C} or {20L, 20C} (both 2 runways, same capacity)
	if len(config) != 2 {
		t.Errorf("Expected 2 active runways, got %d", len(config))
	}

	activeRunways := make([]string, 0, len(config))
	for k := range config {
		activeRunways = append(activeRunways, k)
	}

	valid02 := containsSameElements(activeRunways, []string{"02L", "02C"})
	valid20 := containsSameElements(activeRunways, []string{"20L", "20C"})

	if !valid02 && !valid20 {
		t.Errorf("Expected {02L, 02C} or {20L, 20C}, got %v", activeRunways)
	}
}

// Test 11: Verify Maximal Cliques Computation
// Test that the Bron-Kerbosch algorithm correctly finds all maximal cliques
func TestRunwayManager_Compatibility_MaximalCliques(t *testing.T) {
	runways := []airport.Runway{
		{RunwayDesignation: "A", TrueBearing: 0, MinimumSeparation: 90 * time.Second},
		{RunwayDesignation: "B", TrueBearing: 90, MinimumSeparation: 90 * time.Second},
		{RunwayDesignation: "C", TrueBearing: 180, MinimumSeparation: 90 * time.Second},
		{RunwayDesignation: "D", TrueBearing: 270, MinimumSeparation: 90 * time.Second},
	}

	// Triangle graph: A-B-C with D isolated
	// Maximal cliques should be: {A, B}, {B, C}, {D}
	compat := airport.NewRunwayCompatibility(map[string][]string{
		"A": {"B"},
		"B": {"A", "C"},
		"C": {"B"},
		"D": {},
	})

	rm := NewRunwayManager(runways, compat)

	// Access the internal state (this is a test, we can do this)
	// Force computation of maximal cliques
	rm.mu.Lock()
	rm.computeMaximalCliques()
	cliques := rm.maximalCliques
	rm.mu.Unlock()

	// Should find 3 maximal cliques
	if len(cliques) != 3 {
		t.Errorf("Expected 3 maximal cliques, got %d: %v", len(cliques), cliques)
	}

	// Verify each expected clique exists
	expectedCliques := [][]string{
		{"A", "B"},
		{"B", "C"},
		{"D"},
	}

	for _, expected := range expectedCliques {
		found := false
		for _, clique := range cliques {
			if containsSameElements(clique, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find clique %v in %v", expected, cliques)
		}
	}
}

// Test 12: Thread Safety During Configuration Changes
// Verify that concurrent availability changes don't break compatibility logic
func TestRunwayManager_Compatibility_ThreadSafety(t *testing.T) {
	runways := []airport.Runway{
		{RunwayDesignation: "09L", TrueBearing: 90, MinimumSeparation: 90 * time.Second},
		{RunwayDesignation: "09R", TrueBearing: 90, MinimumSeparation: 90 * time.Second},
		{RunwayDesignation: "18", TrueBearing: 180, MinimumSeparation: 90 * time.Second},
	}

	compat := airport.NewRunwayCompatibility(map[string][]string{
		"09L": {"09R"},
		"09R": {"09L"},
		"18":  {},
	})

	rm := NewRunwayManager(runways, compat)

	// Concurrent operations
	done := make(chan bool, 3)

	// Reader
	go func() {
		for i := 0; i < 100; i++ {
			config := rm.GetActiveConfiguration()
			if config == nil {
				t.Error("Got nil configuration")
			}
		}
		done <- true
	}()

	// Writer 1
	go func() {
		for i := 0; i < 50; i++ {
			rm.OnRunwayUnavailable("09L")
			rm.OnRunwayAvailable("09L")
		}
		done <- true
	}()

	// Writer 2
	go func() {
		for i := 0; i < 50; i++ {
			rm.OnRunwayUnavailable("18")
			rm.OnRunwayAvailable("18")
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done

	// Final config should be valid
	finalConfig := rm.GetActiveConfiguration()
	if finalConfig == nil {
		t.Error("Final configuration should not be nil")
	}
}
