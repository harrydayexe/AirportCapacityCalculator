package policy

import (
	"context"
	"testing"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/event"
)

func TestNewGateCapacityPolicy(t *testing.T) {
	tests := []struct {
		name        string
		constraint  GateCapacityConstraint
		expectError bool
	}{
		{
			name: "valid constraint",
			constraint: GateCapacityConstraint{
				TotalGates:            50,
				AverageTurnaroundTime: 2 * time.Hour,
			},
			expectError: false,
		},
		{
			name: "zero gates",
			constraint: GateCapacityConstraint{
				TotalGates:            0,
				AverageTurnaroundTime: 2 * time.Hour,
			},
			expectError: true,
		},
		{
			name: "negative gates",
			constraint: GateCapacityConstraint{
				TotalGates:            -10,
				AverageTurnaroundTime: 2 * time.Hour,
			},
			expectError: true,
		},
		{
			name: "zero turnaround time",
			constraint: GateCapacityConstraint{
				TotalGates:            50,
				AverageTurnaroundTime: 0,
			},
			expectError: true,
		},
		{
			name: "negative turnaround time",
			constraint: GateCapacityConstraint{
				TotalGates:            50,
				AverageTurnaroundTime: -1 * time.Hour,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := NewGateCapacityPolicy(tt.constraint)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if policy != nil {
					t.Error("Expected nil policy on error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if policy == nil {
					t.Error("Expected non-nil policy")
				}
			}
		})
	}
}

func TestGateCapacityPolicy_Name(t *testing.T) {
	policy, _ := NewGateCapacityPolicy(GateCapacityConstraint{
		TotalGates:            50,
		AverageTurnaroundTime: 2 * time.Hour,
	})

	if policy.Name() != "GateCapacityPolicy" {
		t.Errorf("Expected policy name 'GateCapacityPolicy', got '%s'", policy.Name())
	}
}

func TestGateCapacityPolicy_GenerateEvents(t *testing.T) {
	simStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	simEnd := simStart.AddDate(0, 0, 7)

	tests := []struct {
		name                      string
		constraint                GateCapacityConstraint
		expectedMovementsPerHour  float32
		tolerance                 float32
	}{
		{
			name: "50 gates, 2 hour turnaround",
			constraint: GateCapacityConstraint{
				TotalGates:            50,
				AverageTurnaroundTime: 2 * time.Hour,
			},
			// 50 gates / 2 hours = 25 arrivals/hour
			// Total movements = 25 * 2 = 50 movements/hour
			expectedMovementsPerHour: 50,
			tolerance:                0.01,
		},
		{
			name: "100 gates, 1 hour turnaround",
			constraint: GateCapacityConstraint{
				TotalGates:            100,
				AverageTurnaroundTime: 1 * time.Hour,
			},
			// 100 gates / 1 hour = 100 arrivals/hour
			// Total movements = 100 * 2 = 200 movements/hour
			expectedMovementsPerHour: 200,
			tolerance:                0.01,
		},
		{
			name: "30 gates, 3 hour turnaround",
			constraint: GateCapacityConstraint{
				TotalGates:            30,
				AverageTurnaroundTime: 3 * time.Hour,
			},
			// 30 gates / 3 hours = 10 arrivals/hour
			// Total movements = 10 * 2 = 20 movements/hour
			expectedMovementsPerHour: 20,
			tolerance:                0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := NewGateCapacityPolicy(tt.constraint)
			if err != nil {
				t.Fatalf("Failed to create policy: %v", err)
			}

			world := newMockEventWorld(simStart, simEnd, []string{"09L", "09R"})
			err = policy.GenerateEvents(context.Background(), world)
			if err != nil {
				t.Fatalf("GenerateEvents failed: %v", err)
			}

			// Should generate exactly one gate capacity constraint event
			gateEvents := world.CountEventsByType(event.GateCapacityConstraintType)
			if gateEvents != 1 {
				t.Errorf("Expected 1 gate capacity event, got %d", gateEvents)
			}

			// Verify the constraint value
			for _, evt := range world.events {
				if evt.Type() == event.GateCapacityConstraintType {
					gateEvt, ok := evt.(*event.GateCapacityConstraintEvent)
					if !ok {
						t.Fatal("Failed to cast event to GateCapacityConstraintEvent")
					}

					// Convert movements per second to movements per hour
					movementsPerHour := gateEvt.MaxMovementsPerSecond() * 3600

					diff := movementsPerHour - tt.expectedMovementsPerHour
					if diff < 0 {
						diff = -diff
					}

					if diff > tt.tolerance {
						t.Errorf("Expected ~%.2f movements/hour, got %.2f (diff: %.2f)",
							tt.expectedMovementsPerHour, movementsPerHour, diff)
					}

					// Event should be at simulation start
					if !evt.Time().Equal(simStart) {
						t.Errorf("Expected event at %v, got %v", simStart, evt.Time())
					}
				}
			}
		})
	}
}

func TestGateCapacityPolicy_IntegrationWithWorld(t *testing.T) {
	simStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	simEnd := simStart.AddDate(0, 0, 1)

	constraint := GateCapacityConstraint{
		TotalGates:            50,
		AverageTurnaroundTime: 2 * time.Hour,
	}

	policy, err := NewGateCapacityPolicy(constraint)
	if err != nil {
		t.Fatalf("Failed to create policy: %v", err)
	}

	world := newMockEventWorld(simStart, simEnd, []string{"09L"})
	err = policy.GenerateEvents(context.Background(), world)
	if err != nil {
		t.Fatalf("GenerateEvents failed: %v", err)
	}

	// Verify the gate capacity event
	foundGateEvent := false
	for _, evt := range world.events {
		if evt.Type() == event.GateCapacityConstraintType {
			foundGateEvent = true

			// The event should have the Apply method that sets the constraint
			gateEvt := evt.(*event.GateCapacityConstraintEvent)
			if gateEvt.MaxMovementsPerSecond() <= 0 {
				t.Error("Expected positive gate capacity constraint")
			}
		}
	}

	if !foundGateEvent {
		t.Error("Expected gate capacity event to be generated")
	}
}
