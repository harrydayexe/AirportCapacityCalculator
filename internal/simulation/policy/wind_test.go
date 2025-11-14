package policy

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/event"
)

// TestNewWindPolicy tests the wind policy constructor
func TestNewWindPolicy(t *testing.T) {
	tests := []struct {
		name          string
		speed         float64
		direction     float64
		expectError   bool
		expectedSpeed float64
		expectedDir   float64
	}{
		{
			name:          "valid wind - northerly",
			speed:         10,
			direction:     360,
			expectError:   false,
			expectedSpeed: 10,
			expectedDir:   0, // 360 normalizes to 0
		},
		{
			name:          "valid wind - easterly",
			speed:         15,
			direction:     90,
			expectError:   false,
			expectedSpeed: 15,
			expectedDir:   90,
		},
		{
			name:          "valid wind - calm",
			speed:         0,
			direction:     0,
			expectError:   false,
			expectedSpeed: 0,
			expectedDir:   0,
		},
		{
			name:          "negative speed",
			speed:         -5,
			direction:     180,
			expectError:   true,
			expectedSpeed: 0,
			expectedDir:   0,
		},
		{
			name:          "direction > 360",
			speed:         20,
			direction:     450, // Should normalize to 90
			expectError:   false,
			expectedSpeed: 20,
			expectedDir:   90,
		},
		{
			name:          "negative direction",
			speed:         10,
			direction:     -90, // Should normalize to 270
			expectError:   false,
			expectedSpeed: 10,
			expectedDir:   270,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := NewWindPolicy(tt.speed, tt.direction)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if policy.GetSpeed() != tt.expectedSpeed {
				t.Errorf("Expected speed %f, got %f", tt.expectedSpeed, policy.GetSpeed())
			}

			if policy.GetDirection() != tt.expectedDir {
				t.Errorf("Expected direction %f, got %f", tt.expectedDir, policy.GetDirection())
			}
		})
	}
}

// TestCalculateWindComponents tests wind component calculations
func TestCalculateWindComponents(t *testing.T) {
	tests := []struct {
		name            string
		runwayBearing   float64
		windSpeed       float64
		windDirection   float64
		expectedHeadwind float64
		expectedCrosswind float64
		tolerance       float64
	}{
		{
			name:            "direct headwind - runway 09, wind 090",
			runwayBearing:   90,
			windSpeed:       20,
			windDirection:   90,
			expectedHeadwind: 20,
			expectedCrosswind: 0,
			tolerance:       0.01,
		},
		{
			name:            "direct tailwind - runway 09, wind 270",
			runwayBearing:   90,
			windSpeed:       20,
			windDirection:   270,
			expectedHeadwind: -20,
			expectedCrosswind: 0,
			tolerance:       0.01,
		},
		{
			name:            "direct crosswind - runway 09, wind 360",
			runwayBearing:   90,
			windSpeed:       20,
			windDirection:   0, // North
			expectedHeadwind: 0,
			expectedCrosswind: 20,
			tolerance:       0.01,
		},
		{
			name:            "direct crosswind - runway 09, wind 180",
			runwayBearing:   90,
			windSpeed:       20,
			windDirection:   180, // South
			expectedHeadwind: 0,
			expectedCrosswind: 20,
			tolerance:       0.01,
		},
		{
			name:            "30 degree angle - runway 09, wind 120",
			runwayBearing:   90,
			windSpeed:       20,
			windDirection:   120,
			expectedHeadwind: 17.32, // 20 * cos(30°)
			expectedCrosswind: 10,    // 20 * sin(30°)
			tolerance:       0.01,
		},
		{
			name:            "45 degree angle - runway 27, wind 315",
			runwayBearing:   270,
			windSpeed:       20,
			windDirection:   315,
			expectedHeadwind: 14.14, // 20 * cos(45°)
			expectedCrosswind: 14.14, // 20 * sin(45°)
			tolerance:       0.01,
		},
		{
			name:            "calm wind",
			runwayBearing:   180,
			windSpeed:       0,
			windDirection:   0,
			expectedHeadwind: 0,
			expectedCrosswind: 0,
			tolerance:       0.01,
		},
		{
			name:            "runway 36, wind 270 (westerly)",
			runwayBearing:   360,
			windSpeed:       15,
			windDirection:   270,
			expectedHeadwind: 0,
			expectedCrosswind: 15,
			tolerance:       0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headwind, crosswind := CalculateWindComponents(tt.runwayBearing, tt.windSpeed, tt.windDirection)

			if math.Abs(headwind-tt.expectedHeadwind) > tt.tolerance {
				t.Errorf("Headwind: expected %f, got %f (diff: %f)",
					tt.expectedHeadwind, headwind, math.Abs(headwind-tt.expectedHeadwind))
			}

			if math.Abs(crosswind-tt.expectedCrosswind) > tt.tolerance {
				t.Errorf("Crosswind: expected %f, got %f (diff: %f)",
					tt.expectedCrosswind, crosswind, math.Abs(crosswind-tt.expectedCrosswind))
			}

			// Crosswind should always be positive
			if crosswind < 0 {
				t.Errorf("Crosswind should always be positive, got %f", crosswind)
			}
		})
	}
}

// TestIsRunwayUsableInWind tests runway usability checks
func TestIsRunwayUsableInWind(t *testing.T) {
	tests := []struct {
		name            string
		windSpeed       float64
		windDirection   float64
		runwayBearing   float64
		crosswindLimit  float64
		tailwindLimit   float64
		expectedUsable  bool
	}{
		{
			name:           "usable - direct headwind within limits",
			windSpeed:      20,
			windDirection:  90,
			runwayBearing:  90,
			crosswindLimit: 35,
			tailwindLimit:  10,
			expectedUsable: true,
		},
		{
			name:           "unusable - excessive tailwind",
			windSpeed:      20,
			windDirection:  270,
			runwayBearing:  90,
			crosswindLimit: 35,
			tailwindLimit:  10,
			expectedUsable: false, // 20kt tailwind exceeds 10kt limit
		},
		{
			name:           "unusable - excessive crosswind",
			windSpeed:      40,
			windDirection:  0,
			runwayBearing:  90,
			crosswindLimit: 35,
			tailwindLimit:  10,
			expectedUsable: false, // 40kt crosswind exceeds 35kt limit
		},
		{
			name:           "usable - at crosswind limit",
			windSpeed:      35,
			windDirection:  0,
			runwayBearing:  90,
			crosswindLimit: 35,
			tailwindLimit:  10,
			expectedUsable: true, // Exactly at limit
		},
		{
			name:           "usable - at tailwind limit",
			windSpeed:      10,
			windDirection:  270,
			runwayBearing:  90,
			crosswindLimit: 35,
			tailwindLimit:  10,
			expectedUsable: true, // Exactly at limit
		},
		{
			name:           "usable - no limits set",
			windSpeed:      50,
			windDirection:  270,
			runwayBearing:  90,
			crosswindLimit: 0, // No limit
			tailwindLimit:  0, // No limit
			expectedUsable: true,
		},
		{
			name:           "usable - calm wind",
			windSpeed:      0,
			windDirection:  0,
			runwayBearing:  90,
			crosswindLimit: 35,
			tailwindLimit:  10,
			expectedUsable: true,
		},
		{
			name:           "usable - slight tailwind within limit",
			windSpeed:      8,
			windDirection:  270, // Direct tailwind of 8kt, within 10kt limit
			runwayBearing:  90,
			crosswindLimit: 35,
			tailwindLimit:  10,
			expectedUsable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := NewWindPolicy(tt.windSpeed, tt.windDirection)
			if err != nil {
				t.Fatalf("Failed to create wind policy: %v", err)
			}

			usable := policy.IsRunwayUsableInWind(tt.runwayBearing, tt.crosswindLimit, tt.tailwindLimit)

			if usable != tt.expectedUsable {
				headwind, crosswind := CalculateWindComponents(tt.runwayBearing, tt.windSpeed, tt.windDirection)
				t.Errorf("Expected usable=%v, got %v (headwind=%.2f, crosswind=%.2f)",
					tt.expectedUsable, usable, headwind, crosswind)
			}
		})
	}
}

// TestWindPolicyName tests the policy name
func TestWindPolicyName(t *testing.T) {
	policy, err := NewWindPolicy(10, 270)
	if err != nil {
		t.Fatalf("Failed to create policy: %v", err)
	}

	if policy.Name() != "WindPolicy" {
		t.Errorf("Expected name 'WindPolicy', got '%s'", policy.Name())
	}
}

// mockWorldState is a mock implementation of WorldState for testing
type mockWorldState struct {
	windSpeed     float64
	windDirection float64
	setWindCalled bool
}

func (m *mockWorldState) SetWind(speed, direction float64) error {
	m.windSpeed = speed
	m.windDirection = direction
	m.setWindCalled = true
	return nil
}

func (m *mockWorldState) ScheduleEvent(evt event.Event)     {}
func (m *mockWorldState) GetEventQueue() *event.EventQueue  { return event.NewEventQueue() }
func (m *mockWorldState) GetStartTime() time.Time           { return time.Time{} }
func (m *mockWorldState) GetEndTime() time.Time             { return time.Time{} }
func (m *mockWorldState) GetRunwayIDs() []string            { return nil }

// TestWindPolicyGenerateEvents tests event generation
func TestWindPolicyGenerateEvents(t *testing.T) {
	policy, err := NewWindPolicy(15, 270)
	if err != nil {
		t.Fatalf("Failed to create policy: %v", err)
	}

	mockWorld := &mockWorldState{}
	ctx := context.Background()

	err = policy.GenerateEvents(ctx, mockWorld)
	if err != nil {
		t.Errorf("GenerateEvents failed: %v", err)
	}

	if !mockWorld.setWindCalled {
		t.Error("SetWind was not called")
	}

	if mockWorld.windSpeed != 15 {
		t.Errorf("Expected wind speed 15, got %f", mockWorld.windSpeed)
	}

	if mockWorld.windDirection != 270 {
		t.Errorf("Expected wind direction 270, got %f", mockWorld.windDirection)
	}
}

// TestWindComponentsReciprocal tests that reciprocal runways have opposite headwind components
func TestWindComponentsReciprocal(t *testing.T) {
	// Test that runway 09 and 27 have opposite headwind components for the same wind
	windSpeed := 20.0
	windDirection := 270.0 // Westerly wind

	headwind09, crosswind09 := CalculateWindComponents(90, windSpeed, windDirection)
	headwind27, crosswind27 := CalculateWindComponents(270, windSpeed, windDirection)

	// Headwinds should be opposite
	if math.Abs(headwind09+headwind27) > 0.01 {
		t.Errorf("Reciprocal runways should have opposite headwinds: 09=%f, 27=%f",
			headwind09, headwind27)
	}

	// Crosswinds should be the same (both positive)
	if math.Abs(crosswind09-crosswind27) > 0.01 {
		t.Errorf("Reciprocal runways should have same crosswind: 09=%f, 27=%f",
			crosswind09, crosswind27)
	}

	// For westerly wind on runway 09, should be tailwind (negative)
	if headwind09 >= 0 {
		t.Errorf("Runway 09 with westerly wind should have tailwind, got headwind=%f", headwind09)
	}

	// For westerly wind on runway 27, should be headwind (positive)
	if headwind27 <= 0 {
		t.Errorf("Runway 27 with westerly wind should have headwind, got headwind=%f", headwind27)
	}
}
