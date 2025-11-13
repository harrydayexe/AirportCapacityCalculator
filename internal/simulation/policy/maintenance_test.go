package policy

import (
	"context"
	"testing"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/event"
)

func TestNewMaintenancePolicy(t *testing.T) {
	schedule := MaintenanceSchedule{
		RunwayDesignations: []string{"09L", "09R"},
		Duration:           4 * time.Hour,
		Frequency:          30 * 24 * time.Hour, // Monthly
	}

	policy := NewMaintenancePolicy(schedule)
	if policy == nil {
		t.Fatal("expected non-nil policy")
	}
	if len(policy.schedule.RunwayDesignations) != 2 {
		t.Errorf("expected 2 runways, got %d", len(policy.schedule.RunwayDesignations))
	}
}

func TestMaintenancePolicy_Name(t *testing.T) {
	schedule := MaintenanceSchedule{
		RunwayDesignations: []string{"09L"},
		Duration:           2 * time.Hour,
		Frequency:          7 * 24 * time.Hour,
	}

	policy := NewMaintenancePolicy(schedule)
	expectedName := "MaintenancePolicy"
	if policy.Name() != expectedName {
		t.Errorf("expected name %q, got %q", expectedName, policy.Name())
	}
}

func TestMaintenancePolicy_GenerateEvents(t *testing.T) {
	tests := []struct {
		name                  string
		runways               []string
		duration              time.Duration
		frequency             time.Duration
		simStart              time.Time
		simEnd                time.Time
		expectedStartEvents   int
		expectedEndEvents     int
	}{
		{
			name:                "Monthly maintenance for one runway over one year",
			runways:             []string{"09L"},
			duration:            4 * time.Hour,
			frequency:           30 * 24 * time.Hour, // ~monthly
			simStart:            time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			simEnd:              time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			expectedStartEvents: 12, // 12 months
			expectedEndEvents:   12,
		},
		{
			name:                "Weekly maintenance for two runways over one month",
			runways:             []string{"09L", "09R"},
			duration:            2 * time.Hour,
			frequency:           7 * 24 * time.Hour, // weekly
			simStart:            time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			simEnd:              time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			expectedStartEvents: 8,  // 4 weeks * 2 runways
			expectedEndEvents:   8,
		},
		{
			name:                "Daily maintenance for one runway over one week",
			runways:             []string{"18"},
			duration:            1 * time.Hour,
			frequency:           24 * time.Hour, // daily
			simStart:            time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			simEnd:              time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),
			expectedStartEvents: 7,
			expectedEndEvents:   7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule := MaintenanceSchedule{
				RunwayDesignations: tt.runways,
				Duration:           tt.duration,
				Frequency:          tt.frequency,
			}

			policy := NewMaintenancePolicy(schedule)
			world := newMockEventWorld(tt.simStart, tt.simEnd, tt.runways)

			err := policy.GenerateEvents(context.Background(), world)
			if err != nil {
				t.Fatalf("GenerateEvents failed: %v", err)
			}

			// Count maintenance events
			startEvents := world.CountEventsByType(event.RunwayMaintenanceStartType)
			endEvents := world.CountEventsByType(event.RunwayMaintenanceEndType)

			if startEvents != tt.expectedStartEvents {
				t.Errorf("expected %d maintenance start events, got %d", tt.expectedStartEvents, startEvents)
			}

			if endEvents != tt.expectedEndEvents {
				t.Errorf("expected %d maintenance end events, got %d", tt.expectedEndEvents, endEvents)
			}
		})
	}
}

func TestMaintenancePolicy_GenerateEvents_InvalidRunway(t *testing.T) {
	schedule := MaintenanceSchedule{
		RunwayDesignations: []string{"INVALID"},
		Duration:           2 * time.Hour,
		Frequency:          7 * 24 * time.Hour,
	}

	policy := NewMaintenancePolicy(schedule)

	simStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	simEnd := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	world := newMockEventWorld(simStart, simEnd, []string{"09L", "09R"})

	err := policy.GenerateEvents(context.Background(), world)
	if err == nil {
		t.Error("expected error for invalid runway, got nil")
	}
}
