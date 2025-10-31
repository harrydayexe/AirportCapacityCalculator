package policy

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/airport"
)

// testLogger creates a test logger that discards output
func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

// mockSimulationState provides a test implementation of the state interface
type mockSimulationState struct {
	operatingHours   float32
	availableRunways []airport.Runway
}

func (m *mockSimulationState) GetOperatingHours() float32 {
	return m.operatingHours
}

func (m *mockSimulationState) SetOperatingHours(hours float32) {
	m.operatingHours = hours
}

func (m *mockSimulationState) GetAvailableRunways() []airport.Runway {
	return m.availableRunways
}

func (m *mockSimulationState) SetAvailableRunways(runways []airport.Runway) {
	m.availableRunways = runways
}

func TestNewDefaultRunwayRotationPolicy(t *testing.T) {
	tests := []struct {
		name     string
		strategy RotationStrategy
	}{
		{"NoRotation", NoRotation},
		{"TimeBasedRotation", TimeBasedRotation},
		{"BalancedRotation", BalancedRotation},
		{"NoiseOptimizedRotation", NoiseOptimizedRotation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := NewDefaultRunwayRotationPolicy(tt.strategy)
			if policy == nil {
				t.Fatal("NewDefaultRunwayRotationPolicy returned nil")
			}
			if policy.strategy != tt.strategy {
				t.Errorf("expected strategy %v, got %v", tt.strategy, policy.strategy)
			}
		})
	}
}

func TestRunwayRotationPolicy_Name(t *testing.T) {
	tests := []struct {
		strategy     RotationStrategy
		expectedName string
	}{
		{NoRotation, "RunwayRotationPolicy(NoRotation)"},
		{TimeBasedRotation, "RunwayRotationPolicy(TimeBasedRotation)"},
		{BalancedRotation, "RunwayRotationPolicy(BalancedRotation)"},
		{NoiseOptimizedRotation, "RunwayRotationPolicy(NoiseOptimizedRotation)"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedName, func(t *testing.T) {
			policy := NewDefaultRunwayRotationPolicy(tt.strategy)
			name := policy.Name()
			if name != tt.expectedName {
				t.Errorf("expected name %q, got %q", tt.expectedName, name)
			}
		})
	}
}

func TestRunwayRotationPolicy_Apply_NoRotation(t *testing.T) {
	policy := NewDefaultRunwayRotationPolicy(NoRotation)
	state := &mockSimulationState{
		operatingHours: 8760,
		availableRunways: []airport.Runway{
			{RunwayDesignation: "09L"},
			{RunwayDesignation: "27R"},
		},
	}

	err := policy.Apply(context.Background(), state, testLogger())
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// NoRotation should not modify operating hours (100% efficiency)
	expectedHours := float32(8760)
	if state.GetOperatingHours() != expectedHours {
		t.Errorf("expected operating hours %f, got %f", expectedHours, state.GetOperatingHours())
	}
}

func TestRunwayRotationPolicy_Apply_TimeBasedRotation(t *testing.T) {
	policy := NewDefaultRunwayRotationPolicy(TimeBasedRotation)
	state := &mockSimulationState{
		operatingHours: 8760,
		availableRunways: []airport.Runway{
			{RunwayDesignation: "09L"},
			{RunwayDesignation: "27R"},
		},
	}

	err := policy.Apply(context.Background(), state, testLogger())
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// TimeBasedRotation should apply 95% efficiency (5% reduction)
	expectedHours := float32(8760 * 0.95)
	if state.GetOperatingHours() != expectedHours {
		t.Errorf("expected operating hours %f, got %f", expectedHours, state.GetOperatingHours())
	}
}

func TestRunwayRotationPolicy_Apply_BalancedRotation(t *testing.T) {
	policy := NewDefaultRunwayRotationPolicy(BalancedRotation)
	state := &mockSimulationState{
		operatingHours: 8760,
		availableRunways: []airport.Runway{
			{RunwayDesignation: "09L"},
			{RunwayDesignation: "27R"},
		},
	}

	err := policy.Apply(context.Background(), state, testLogger())
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// BalancedRotation should apply 90% efficiency (10% reduction)
	expectedHours := float32(8760 * 0.90)
	if state.GetOperatingHours() != expectedHours {
		t.Errorf("expected operating hours %f, got %f", expectedHours, state.GetOperatingHours())
	}
}

func TestRunwayRotationPolicy_Apply_NoiseOptimizedRotation(t *testing.T) {
	policy := NewDefaultRunwayRotationPolicy(NoiseOptimizedRotation)
	state := &mockSimulationState{
		operatingHours: 8760,
		availableRunways: []airport.Runway{
			{RunwayDesignation: "09L"},
			{RunwayDesignation: "27R"},
		},
	}

	err := policy.Apply(context.Background(), state, testLogger())
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// NoiseOptimizedRotation should apply 80% efficiency (20% reduction)
	expectedHours := float32(8760 * 0.80)
	if state.GetOperatingHours() != expectedHours {
		t.Errorf("expected operating hours %f, got %f", expectedHours, state.GetOperatingHours())
	}
}

func TestRunwayRotationPolicy_Apply_DefaultOperatingHours(t *testing.T) {
	// Test that policy correctly handles zero operating hours by defaulting to 8760
	policy := NewDefaultRunwayRotationPolicy(BalancedRotation)
	state := &mockSimulationState{
		operatingHours: 0, // Not set, should default to 8760
		availableRunways: []airport.Runway{
			{RunwayDesignation: "09L"},
		},
	}

	err := policy.Apply(context.Background(), state, testLogger())
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Should apply 90% efficiency to default 8760 hours
	expectedHours := float32(8760 * 0.90)
	if state.GetOperatingHours() != expectedHours {
		t.Errorf("expected operating hours %f, got %f", expectedHours, state.GetOperatingHours())
	}
}

func TestRunwayRotationPolicy_Apply_EfficiencyMultipliers(t *testing.T) {
	// Test all strategies to verify the efficiency multiplier values
	tests := []struct {
		strategy             RotationStrategy
		expectedMultiplier   float32
		expectedReduction    float32
		initialHours         float32
	}{
		{NoRotation, 1.0, 0.0, 10000},
		{TimeBasedRotation, 0.95, 0.05, 10000},
		{BalancedRotation, 0.90, 0.10, 10000},
		{NoiseOptimizedRotation, 0.80, 0.20, 10000},
	}

	for _, tt := range tests {
		t.Run(tt.strategy.String(), func(t *testing.T) {
			policy := NewDefaultRunwayRotationPolicy(tt.strategy)
			state := &mockSimulationState{
				operatingHours: tt.initialHours,
				availableRunways: []airport.Runway{
					{RunwayDesignation: "09L"},
				},
			}

			err := policy.Apply(context.Background(), state, testLogger())
			if err != nil {
				t.Fatalf("Apply failed: %v", err)
			}

			expectedHours := tt.initialHours * tt.expectedMultiplier
			if state.GetOperatingHours() != expectedHours {
				t.Errorf("expected operating hours %f (%.0f%% of %f), got %f",
					expectedHours, tt.expectedMultiplier*100, tt.initialHours, state.GetOperatingHours())
			}

			// Verify the reduction percentage
			actualReduction := (tt.initialHours - state.GetOperatingHours()) / tt.initialHours
			if actualReduction != tt.expectedReduction {
				t.Errorf("expected %.0f%% reduction, got %.0f%%", tt.expectedReduction*100, actualReduction*100)
			}
		})
	}
}

func TestRunwayRotationPolicy_Apply_InvalidState(t *testing.T) {
	policy := NewDefaultRunwayRotationPolicy(BalancedRotation)

	// Pass an invalid state type
	err := policy.Apply(context.Background(), "invalid state", testLogger())
	if err == nil {
		t.Fatal("expected error for invalid state type, got nil")
	}

	expectedError := "invalid state type for RunwayRotationPolicy"
	if err.Error() != expectedError {
		t.Errorf("expected error %q, got %q", expectedError, err.Error())
	}
}

func TestRotationStrategy_String(t *testing.T) {
	tests := []struct {
		strategy RotationStrategy
		expected string
	}{
		{NoRotation, "NoRotation"},
		{TimeBasedRotation, "TimeBasedRotation"},
		{BalancedRotation, "BalancedRotation"},
		{NoiseOptimizedRotation, "NoiseOptimizedRotation"},
		{RotationStrategy(999), "Unknown"}, // Invalid strategy
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.strategy.String()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestRunwayRotationPolicy_Apply_CompoundPolicies(t *testing.T) {
	// Test that rotation policy works correctly when applied after another policy
	// that has already modified operating hours (e.g., curfew policy)

	policy := NewDefaultRunwayRotationPolicy(BalancedRotation)

	// Simulate state after a curfew policy has reduced hours from 8760 to 7000
	state := &mockSimulationState{
		operatingHours: 7000,
		availableRunways: []airport.Runway{
			{RunwayDesignation: "09L"},
			{RunwayDesignation: "27R"},
		},
	}

	err := policy.Apply(context.Background(), state, testLogger())
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Should apply 90% efficiency to the already-reduced 7000 hours
	expectedHours := float32(7000 * 0.90)
	if state.GetOperatingHours() != expectedHours {
		t.Errorf("expected operating hours %f, got %f", expectedHours, state.GetOperatingHours())
	}
}
