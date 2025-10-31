package policy

import (
	"context"
	"testing"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/event"
	"time"
)

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
				t.Fatal("expected non-nil policy")
			}
			if policy.strategy != tt.strategy {
				t.Errorf("expected strategy %v, got %v", tt.strategy, policy.strategy)
			}
		})
	}
}

func TestRunwayRotationPolicy_Name(t *testing.T) {
	policy := NewDefaultRunwayRotationPolicy(NoRotation)
	expectedName := "RunwayRotationPolicy(NoRotation)"
	if policy.Name() != expectedName {
		t.Errorf("expected name %q, got %q", expectedName, policy.Name())
	}
}

func TestRunwayRotationPolicy_GenerateEvents(t *testing.T) {
	tests := []struct {
		name               string
		strategy           RotationStrategy
		expectedMultiplier float32
	}{
		{"NoRotation", NoRotation, 1.0},
		{"TimeBasedRotation", TimeBasedRotation, 0.95},
		{"BalancedRotation", BalancedRotation, 0.90},
		{"NoiseOptimizedRotation", NoiseOptimizedRotation, 0.80},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := NewDefaultRunwayRotationPolicy(tt.strategy)

			simStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			simEnd := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
			world := newMockEventWorld(simStart, simEnd, []string{"09L", "09R"})

			err := policy.GenerateEvents(context.Background(), world)
			if err != nil {
				t.Fatalf("GenerateEvents failed: %v", err)
			}

			// Should generate exactly one rotation change event
			rotationEvents := world.CountEventsByType(event.RotationChangeType)
			if rotationEvents != 1 {
				t.Errorf("expected 1 rotation event, got %d", rotationEvents)
			}

			// Verify the event is at the start time
			events := world.GetEvents()
			if len(events) == 0 {
				t.Fatal("expected at least one event")
			}

			rotEvent := events[0]
			if rotEvent.Type() != event.RotationChangeType {
				t.Errorf("expected RotationChange event, got %s", rotEvent.Type())
			}

			if !rotEvent.Time().Equal(simStart) {
				t.Errorf("expected event at sim start (%v), got %v", simStart, rotEvent.Time())
			}

			// Verify the multiplier (cast to concrete type to check)
			if rotChangeEvent, ok := rotEvent.(*event.RotationChangeEvent); ok {
				if rotChangeEvent.Multiplier() != tt.expectedMultiplier {
					t.Errorf("expected multiplier %f, got %f", tt.expectedMultiplier, rotChangeEvent.Multiplier())
				}
			} else {
				t.Error("could not cast event to RotationChangeEvent")
			}
		})
	}
}

func TestRunwayRotationPolicy_CustomConfiguration(t *testing.T) {
	customConfig := NewRotationPolicyConfiguration(map[RotationStrategy]float32{
		NoRotation:             0.99,
		TimeBasedRotation:      0.85,
		BalancedRotation:       0.75,
		NoiseOptimizedRotation: 0.65,
	})

	policy := NewRunwayRotationPolicy(BalancedRotation, customConfig)

	simStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	simEnd := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	world := newMockEventWorld(simStart, simEnd, []string{"09L"})

	err := policy.GenerateEvents(context.Background(), world)
	if err != nil {
		t.Fatalf("GenerateEvents failed: %v", err)
	}

	events := world.GetEvents()
	if len(events) == 0 {
		t.Fatal("expected at least one event")
	}

	if rotChangeEvent, ok := events[0].(*event.RotationChangeEvent); ok {
		expectedMultiplier := float32(0.75)
		if rotChangeEvent.Multiplier() != expectedMultiplier {
			t.Errorf("expected custom multiplier %f, got %f", expectedMultiplier, rotChangeEvent.Multiplier())
		}
	}
}
