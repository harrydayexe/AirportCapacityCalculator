package policy

import (
	"context"
	"testing"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/event"
)

func TestNewTaxiTimePolicy(t *testing.T) {
	tests := []struct {
		name        string
		config      TaxiTimeConfiguration
		expectError bool
	}{
		{
			name: "valid configuration",
			config: TaxiTimeConfiguration{
				AverageTaxiInTime:  5 * time.Minute,
				AverageTaxiOutTime: 5 * time.Minute,
			},
			expectError: false,
		},
		{
			name: "zero taxi times",
			config: TaxiTimeConfiguration{
				AverageTaxiInTime:  0,
				AverageTaxiOutTime: 0,
			},
			expectError: false,
		},
		{
			name: "negative taxi-in time",
			config: TaxiTimeConfiguration{
				AverageTaxiInTime:  -5 * time.Minute,
				AverageTaxiOutTime: 5 * time.Minute,
			},
			expectError: true,
		},
		{
			name: "negative taxi-out time",
			config: TaxiTimeConfiguration{
				AverageTaxiInTime:  5 * time.Minute,
				AverageTaxiOutTime: -5 * time.Minute,
			},
			expectError: true,
		},
		{
			name: "asymmetric taxi times",
			config: TaxiTimeConfiguration{
				AverageTaxiInTime:  10 * time.Minute,
				AverageTaxiOutTime: 5 * time.Minute,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := NewTaxiTimePolicy(tt.config)
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

func TestTaxiTimePolicy_Name(t *testing.T) {
	policy, _ := NewTaxiTimePolicy(TaxiTimeConfiguration{
		AverageTaxiInTime:  5 * time.Minute,
		AverageTaxiOutTime: 5 * time.Minute,
	})

	if policy.Name() != "TaxiTimePolicy" {
		t.Errorf("Expected policy name 'TaxiTimePolicy', got '%s'", policy.Name())
	}
}

func TestTaxiTimePolicy_GenerateEvents(t *testing.T) {
	simStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	simEnd := simStart.AddDate(0, 0, 7)

	tests := []struct {
		name                    string
		config                  TaxiTimeConfiguration
		expectedTotalOverhead   time.Duration
	}{
		{
			name: "5 min in, 5 min out",
			config: TaxiTimeConfiguration{
				AverageTaxiInTime:  5 * time.Minute,
				AverageTaxiOutTime: 5 * time.Minute,
			},
			expectedTotalOverhead: 10 * time.Minute,
		},
		{
			name: "10 min in, 8 min out",
			config: TaxiTimeConfiguration{
				AverageTaxiInTime:  10 * time.Minute,
				AverageTaxiOutTime: 8 * time.Minute,
			},
			expectedTotalOverhead: 18 * time.Minute,
		},
		{
			name: "zero overhead",
			config: TaxiTimeConfiguration{
				AverageTaxiInTime:  0,
				AverageTaxiOutTime: 0,
			},
			expectedTotalOverhead: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := NewTaxiTimePolicy(tt.config)
			if err != nil {
				t.Fatalf("Failed to create policy: %v", err)
			}

			world := newMockEventWorld(simStart, simEnd, []string{"09L"})
			err = policy.GenerateEvents(context.Background(), world)
			if err != nil {
				t.Fatalf("GenerateEvents failed: %v", err)
			}

			// Should generate exactly one taxi time adjustment event
			taxiEvents := world.CountEventsByType(event.TaxiTimeAdjustmentType)
			if taxiEvents != 1 {
				t.Errorf("Expected 1 taxi time event, got %d", taxiEvents)
			}

			// Verify the overhead value
			for _, evt := range world.events {
				if evt.Type() == event.TaxiTimeAdjustmentType {
					taxiEvt, ok := evt.(*event.TaxiTimeAdjustmentEvent)
					if !ok {
						t.Fatal("Failed to cast event to TaxiTimeAdjustmentEvent")
					}

					if taxiEvt.TotalTaxiTimeOverhead() != tt.expectedTotalOverhead {
						t.Errorf("Expected %v overhead, got %v",
							tt.expectedTotalOverhead, taxiEvt.TotalTaxiTimeOverhead())
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

func TestTaxiTimePolicy_Integration(t *testing.T) {
	simStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	simEnd := simStart.AddDate(0, 0, 1)

	config := TaxiTimeConfiguration{
		AverageTaxiInTime:  7 * time.Minute,
		AverageTaxiOutTime: 8 * time.Minute,
	}

	policy, err := NewTaxiTimePolicy(config)
	if err != nil {
		t.Fatalf("Failed to create policy: %v", err)
	}

	world := newMockEventWorld(simStart, simEnd, []string{"09L", "09R"})
	err = policy.GenerateEvents(context.Background(), world)
	if err != nil {
		t.Fatalf("GenerateEvents failed: %v", err)
	}

	// Verify the taxi time event exists
	foundTaxiEvent := false
	for _, evt := range world.events {
		if evt.Type() == event.TaxiTimeAdjustmentType {
			foundTaxiEvent = true

			taxiEvt := evt.(*event.TaxiTimeAdjustmentEvent)
			expectedOverhead := 15 * time.Minute // 7 + 8

			if taxiEvt.TotalTaxiTimeOverhead() != expectedOverhead {
				t.Errorf("Expected %v total overhead, got %v",
					expectedOverhead, taxiEvt.TotalTaxiTimeOverhead())
			}
		}
	}

	if !foundTaxiEvent {
		t.Error("Expected taxi time event to be generated")
	}
}
