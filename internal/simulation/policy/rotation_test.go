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
		{"PreferentialRunway", PreferentialRunway},
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
		{"PreferentialRunway", PreferentialRunway, 0.90},
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
		PreferentialRunway:     0.75,
		NoiseOptimizedRotation: 0.65,
	})

	policy := NewRunwayRotationPolicy(PreferentialRunway, customConfig)

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

func TestRunwayRotationPolicy_TimeBoundedSchedule(t *testing.T) {
	// Test time-bounded rotation (weekends only, 6 AM - 11 PM)
	schedule := &RotationSchedule{
		StartHour:  6,  // 6 AM
		EndHour:    23, // 11 PM
		DaysOfWeek: []time.Weekday{time.Saturday, time.Sunday},
	}

	config := NewDefaultRotationPolicyConfiguration()
	policy := NewRunwayRotationPolicyWithSchedule(TimeBasedRotation, config, schedule)

	// Simulate 2 weeks
	simStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) // Monday
	simEnd := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	world := newMockEventWorld(simStart, simEnd, []string{"09L"})

	err := policy.GenerateEvents(context.Background(), world)
	if err != nil {
		t.Fatalf("GenerateEvents failed: %v", err)
	}

	// Should generate events for 2 weekends:
	// - Saturday Jan 6: start (6 AM), end (11 PM)
	// - Sunday Jan 7: start (6 AM), end (11 PM)
	// - Saturday Jan 13: start (6 AM), end (11 PM)
	// - Sunday Jan 14: start (6 AM), end (11 PM)
	// Total: 8 events
	rotationEvents := world.CountEventsByType(event.RotationChangeType)
	expectedEvents := 8
	if rotationEvents != expectedEvents {
		t.Errorf("expected %d rotation events, got %d", expectedEvents, rotationEvents)
	}

	events := world.GetEvents()

	// Verify alternating pattern: 0.95 (start) -> 1.0 (end) -> 0.95 (start) -> 1.0 (end)...
	expectedMultipliers := []float32{0.95, 1.0, 0.95, 1.0, 0.95, 1.0, 0.95, 1.0}
	for i, expectedMult := range expectedMultipliers {
		if i >= len(events) {
			t.Fatalf("not enough events: expected at least %d, got %d", i+1, len(events))
		}

		rotEvent, ok := events[i].(*event.RotationChangeEvent)
		if !ok {
			t.Errorf("event %d is not a RotationChangeEvent", i)
			continue
		}

		if rotEvent.Multiplier() != expectedMult {
			t.Errorf("event %d: expected multiplier %f, got %f", i, expectedMult, rotEvent.Multiplier())
		}
	}
}

func TestRunwayRotationPolicy_TimeBoundedSchedule_AllDays(t *testing.T) {
	// Test time-bounded rotation applied every day (9 AM - 5 PM)
	schedule := &RotationSchedule{
		StartHour:  9,  // 9 AM
		EndHour:    17, // 5 PM
		DaysOfWeek: nil, // All days
	}

	config := NewDefaultRotationPolicyConfiguration()
	policy := NewRunwayRotationPolicyWithSchedule(PreferentialRunway, config, schedule)

	// Simulate 3 days
	simStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	simEnd := time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC)
	world := newMockEventWorld(simStart, simEnd, []string{"09L"})

	err := policy.GenerateEvents(context.Background(), world)
	if err != nil {
		t.Fatalf("GenerateEvents failed: %v", err)
	}

	// Should generate events for 3 days:
	// - Jan 1: start (9 AM), end (5 PM)
	// - Jan 2: start (9 AM), end (5 PM)
	// - Jan 3: start (9 AM), end (5 PM)
	// Total: 6 events
	rotationEvents := world.CountEventsByType(event.RotationChangeType)
	expectedEvents := 6
	if rotationEvents != expectedEvents {
		t.Errorf("expected %d rotation events, got %d", expectedEvents, rotationEvents)
	}

	events := world.GetEvents()

	// Verify the first event is at 9 AM on Jan 1 with multiplier 0.90 (PreferentialRunway)
	if len(events) > 0 {
		firstEvent, ok := events[0].(*event.RotationChangeEvent)
		if !ok {
			t.Fatal("first event is not a RotationChangeEvent")
		}

		expectedTime := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
		if !firstEvent.Time().Equal(expectedTime) {
			t.Errorf("first event time: expected %v, got %v", expectedTime, firstEvent.Time())
		}

		expectedMult := float32(0.90)
		if firstEvent.Multiplier() != expectedMult {
			t.Errorf("first event multiplier: expected %f, got %f", expectedMult, firstEvent.Multiplier())
		}
	}
}
