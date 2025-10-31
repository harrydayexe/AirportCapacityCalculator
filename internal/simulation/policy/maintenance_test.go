package policy

import (
	"context"
	"testing"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/airport"
)

func TestNewMaintenancePolicy(t *testing.T) {
	schedule := MaintenanceSchedule{
		RunwayDesignations: []string{"09L", "27R"},
		Duration:           4 * time.Hour,
		Frequency:          30 * 24 * time.Hour, // Every 30 days
	}

	policy := NewMaintenancePolicy(schedule)
	if policy == nil {
		t.Fatal("NewMaintenancePolicy returned nil")
	}
	if len(policy.schedule.RunwayDesignations) != 2 {
		t.Errorf("expected 2 runway designations, got %d", len(policy.schedule.RunwayDesignations))
	}
	if policy.schedule.Duration != 4*time.Hour {
		t.Errorf("expected duration 4h, got %v", policy.schedule.Duration)
	}
}

func TestMaintenancePolicy_Name(t *testing.T) {
	schedule := MaintenanceSchedule{
		RunwayDesignations: []string{"09L"},
		Duration:           2 * time.Hour,
		Frequency:          24 * time.Hour,
	}

	policy := NewMaintenancePolicy(schedule)
	name := policy.Name()
	expectedName := "MaintenancePolicy"
	if name != expectedName {
		t.Errorf("expected name %q, got %q", expectedName, name)
	}
}

func TestMaintenancePolicy_Apply(t *testing.T) {
	tests := []struct {
		name                     string
		runwayDesignations       []string
		duration                 time.Duration
		frequency                time.Duration
		initialHours             float32
		expectedMaintenanceHours float64
	}{
		{
			name:                     "Monthly 4-hour maintenance",
			runwayDesignations:       []string{"09L"},
			duration:                 4 * time.Hour,
			frequency:                30 * 24 * time.Hour, // Every 30 days
			initialHours:             8760,
			expectedMaintenanceHours: (365.0 / 30.0) * 4.0, // ~12 windows * 4 hours
		},
		{
			name:                     "Weekly 2-hour maintenance",
			runwayDesignations:       []string{"09L", "27R"},
			duration:                 2 * time.Hour,
			frequency:                7 * 24 * time.Hour, // Every 7 days
			initialHours:             8760,
			expectedMaintenanceHours: (365.0 / 7.0) * 2.0, // ~52 windows * 2 hours
		},
		{
			name:                     "Quarterly 8-hour maintenance",
			runwayDesignations:       []string{"18"},
			duration:                 8 * time.Hour,
			frequency:                90 * 24 * time.Hour, // Every 90 days
			initialHours:             8760,
			expectedMaintenanceHours: (365.0 / 90.0) * 8.0, // ~4 windows * 8 hours
		},
		{
			name:                     "Daily 1-hour maintenance",
			runwayDesignations:       []string{"09L"},
			duration:                 1 * time.Hour,
			frequency:                24 * time.Hour, // Daily
			initialHours:             8760,
			expectedMaintenanceHours: 365.0 * 1.0, // 365 windows * 1 hour
		},
		{
			name:                     "Zero initial hours defaults to full year",
			runwayDesignations:       []string{"09L"},
			duration:                 4 * time.Hour,
			frequency:                30 * 24 * time.Hour,
			initialHours:             0, // Should default to 8760
			expectedMaintenanceHours: (365.0 / 30.0) * 4.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule := MaintenanceSchedule{
				RunwayDesignations: tt.runwayDesignations,
				Duration:           tt.duration,
				Frequency:          tt.frequency,
			}
			policy := NewMaintenancePolicy(schedule)

			state := &mockSimulationState{
				operatingHours: tt.initialHours,
				availableRunways: []airport.Runway{
					{RunwayDesignation: "09L"},
					{RunwayDesignation: "27R"},
				},
			}

			initialOperatingHours := tt.initialHours
			if initialOperatingHours == 0 {
				initialOperatingHours = 8760 // default
			}

			err := policy.Apply(context.Background(), state, testLogger())
			if err != nil {
				t.Fatalf("Apply failed: %v", err)
			}

			// Verify that operating hours were reduced
			expectedHours := initialOperatingHours - float32(tt.expectedMaintenanceHours)
			if expectedHours < 0 {
				expectedHours = 0
			}

			// Allow for small floating point differences
			diff := state.operatingHours - expectedHours
			if diff < -1.0 || diff > 1.0 { // Allow 1 hour tolerance for rounding
				t.Errorf("expected operating hours ~%v, got %v (maintenance hours: %v)",
					expectedHours, state.operatingHours, tt.expectedMaintenanceHours)
			}

			// Ensure hours never go negative
			if state.operatingHours < 0 {
				t.Errorf("operating hours should never be negative, got %v", state.operatingHours)
			}

			// Verify hours were actually reduced (unless already 0)
			if initialOperatingHours > 0 && state.operatingHours >= initialOperatingHours {
				t.Errorf("operating hours should be reduced: before=%v, after=%v",
					initialOperatingHours, state.operatingHours)
			}
		})
	}
}

func TestMaintenancePolicy_Apply_InvalidState(t *testing.T) {
	schedule := MaintenanceSchedule{
		RunwayDesignations: []string{"09L"},
		Duration:           2 * time.Hour,
		Frequency:          24 * time.Hour,
	}
	policy := NewMaintenancePolicy(schedule)

	// Test with an invalid state type
	err := policy.Apply(context.Background(), "invalid state", testLogger())
	if err == nil {
		t.Error("expected error for invalid state type, got nil")
	}
}

func TestMaintenancePolicy_Apply_HighFrequency(t *testing.T) {
	// Test extreme case: maintenance that would exceed available hours
	schedule := MaintenanceSchedule{
		RunwayDesignations: []string{"09L"},
		Duration:           12 * time.Hour, // 12 hour maintenance
		Frequency:          24 * time.Hour, // Every day
	}
	policy := NewMaintenancePolicy(schedule)

	state := &mockSimulationState{
		operatingHours: 8760, // Full year
	}

	err := policy.Apply(context.Background(), state, testLogger())
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Should reduce by 12 hours * 365 days = 4380 hours
	expectedHours := float32(8760 - 4380)

	diff := state.operatingHours - expectedHours
	if diff < -1.0 || diff > 1.0 {
		t.Errorf("expected operating hours ~%v, got %v", expectedHours, state.operatingHours)
	}

	// Ensure result is never negative
	if state.operatingHours < 0 {
		t.Errorf("operating hours should never be negative, got %v", state.operatingHours)
	}
}

func TestMaintenancePolicy_Apply_MultipleRunways(t *testing.T) {
	// Test with multiple runways specified
	schedule := MaintenanceSchedule{
		RunwayDesignations: []string{"09L", "09R", "18", "27L", "27R"},
		Duration:           3 * time.Hour,
		Frequency:          14 * 24 * time.Hour, // Bi-weekly
	}
	policy := NewMaintenancePolicy(schedule)

	state := &mockSimulationState{
		operatingHours: 8760,
		availableRunways: []airport.Runway{
			{RunwayDesignation: "09L"},
			{RunwayDesignation: "09R"},
			{RunwayDesignation: "18"},
			{RunwayDesignation: "27L"},
			{RunwayDesignation: "27R"},
		},
	}

	err := policy.Apply(context.Background(), state, testLogger())
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Verify hours were reduced
	if state.operatingHours >= 8760 {
		t.Errorf("operating hours should be reduced from 8760, got %v", state.operatingHours)
	}

	// Verify hours are reasonable
	expectedMaintenanceHours := (365.0 / 14.0) * 3.0 // ~26 windows * 3 hours
	expectedHours := float32(8760 - expectedMaintenanceHours)

	diff := state.operatingHours - expectedHours
	if diff < -1.0 || diff > 1.0 {
		t.Errorf("expected operating hours ~%v, got %v", expectedHours, state.operatingHours)
	}
}
