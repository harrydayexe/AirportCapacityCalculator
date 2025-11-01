package policy

import (
	"context"
	"testing"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/event"
)

func TestIntelligentMaintenancePolicy_CurfewCoordination(t *testing.T) {
	// Setup: 7-day simulation with nightly curfew (23:00-06:00)
	simStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	simEnd := simStart.AddDate(0, 0, 7)
	curfewStart := time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC)
	curfewEnd := time.Date(2024, 1, 2, 6, 0, 0, 0, time.UTC)

	schedule := IntelligentMaintenanceSchedule{
		RunwayDesignations:       []string{"09L"},
		Duration:                 4 * time.Hour, // 4-hour maintenance fits in 7-hour curfew
		Frequency:                7 * 24 * time.Hour, // Once per week
		MinimumOperationalRunways: 1,
		CurfewStart:              &curfewStart,
		CurfewEnd:                &curfewEnd,
	}

	policy, err := NewIntelligentMaintenancePolicy(schedule)
	if err != nil {
		t.Fatalf("Failed to create policy: %v", err)
	}

	world := newMockEventWorld(simStart, simEnd, []string{"09L"})
	err = policy.GenerateEvents(context.Background(), world)
	if err != nil {
		t.Fatalf("GenerateEvents failed: %v", err)
	}

	// Should schedule maintenance during the curfew
	maintenanceStarts := world.CountEventsByType(event.RunwayMaintenanceStartType)
	if maintenanceStarts == 0 {
		t.Error("Expected at least one maintenance start event")
	}

	// Verify maintenance is scheduled during curfew period
	for _, evt := range world.events {
		if evt.Type() == event.RunwayMaintenanceStartType {
			eventTime := evt.Time()
			hour := eventTime.Hour()
			// Should be during curfew: 23:00-06:00
			if hour < 23 && hour >= 6 {
				t.Errorf("Maintenance scheduled outside curfew: %v (hour: %d)", eventTime, hour)
			}
		}
	}
}

func TestIntelligentMaintenancePolicy_RunwayCoordination(t *testing.T) {
	// Setup: 2 runways, minimum 1 operational
	simStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	simEnd := simStart.AddDate(0, 0, 7)

	schedule := IntelligentMaintenanceSchedule{
		RunwayDesignations:       []string{"09L", "09R"},
		Duration:                 2 * time.Hour,
		Frequency:                24 * time.Hour, // Daily maintenance
		MinimumOperationalRunways: 1, // At least 1 runway must stay operational
	}

	policy, err := NewIntelligentMaintenancePolicy(schedule)
	if err != nil {
		t.Fatalf("Failed to create policy: %v", err)
	}

	world := newMockEventWorld(simStart, simEnd, []string{"09L", "09R"})
	err = policy.GenerateEvents(context.Background(), world)
	if err != nil {
		t.Fatalf("GenerateEvents failed: %v", err)
	}

	// Build maintenance windows map
	maintenanceWindows := make(map[string][]TimeWindow)
	eventsByRunway := make(map[string][]event.Event)

	for _, evt := range world.events {
		switch e := evt.(type) {
		case *event.RunwayMaintenanceStartEvent:
			runwayID := e.RunwayID()
			eventsByRunway[runwayID] = append(eventsByRunway[runwayID], evt)
		case *event.RunwayMaintenanceEndEvent:
			runwayID := e.RunwayID()
			eventsByRunway[runwayID] = append(eventsByRunway[runwayID], evt)
		}
	}

	// Build windows from start/end pairs
	for runwayID, events := range eventsByRunway {
		var startTime time.Time
		for _, evt := range events {
			if evt.Type() == event.RunwayMaintenanceStartType {
				startTime = evt.Time()
			} else if evt.Type() == event.RunwayMaintenanceEndType && !startTime.IsZero() {
				maintenanceWindows[runwayID] = append(maintenanceWindows[runwayID], TimeWindow{
					Start: startTime,
					End:   evt.Time(),
				})
				startTime = time.Time{}
			}
		}
	}

	// Verify no overlap (at least 1 runway always operational)
	for _, window1 := range maintenanceWindows["09L"] {
		for _, window2 := range maintenanceWindows["09R"] {
			// Check if windows overlap
			if window1.Start.Before(window2.End) && window1.End.After(window2.Start) {
				t.Errorf("Maintenance windows overlap: 09L[%v-%v] and 09R[%v-%v]",
					window1.Start, window1.End, window2.Start, window2.End)
			}
		}
	}
}

func TestIntelligentMaintenancePolicy_PeakHoursAvoidance(t *testing.T) {
	// Setup: Avoid peak hours 08:00-20:00
	simStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	simEnd := simStart.AddDate(0, 0, 7)

	schedule := IntelligentMaintenanceSchedule{
		RunwayDesignations:       []string{"09L"},
		Duration:                 2 * time.Hour,
		Frequency:                48 * time.Hour, // Every 2 days
		MinimumOperationalRunways: 1,
		PeakHours: &PeakHours{
			StartHour: 8,
			EndHour:   20,
		},
	}

	policy, err := NewIntelligentMaintenancePolicy(schedule)
	if err != nil {
		t.Fatalf("Failed to create policy: %v", err)
	}

	world := newMockEventWorld(simStart, simEnd, []string{"09L"})
	err = policy.GenerateEvents(context.Background(), world)
	if err != nil {
		t.Fatalf("GenerateEvents failed: %v", err)
	}

	// Verify maintenance starts outside peak hours
	for _, evt := range world.events {
		if evt.Type() == event.RunwayMaintenanceStartType {
			hour := evt.Time().Hour()
			if hour >= 8 && hour < 20 {
				t.Errorf("Maintenance scheduled during peak hours: %v (hour: %d)", evt.Time(), hour)
			}
		}
	}
}

func TestIntelligentMaintenancePolicy_CurfewAdjacent(t *testing.T) {
	// Setup: Maintenance that can be scheduled adjacent to curfew
	simStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	simEnd := simStart.AddDate(0, 0, 3)
	curfewStart := time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC)
	curfewEnd := time.Date(2024, 1, 2, 1, 0, 0, 0, time.UTC) // Short 2-hour curfew

	schedule := IntelligentMaintenanceSchedule{
		RunwayDesignations:       []string{"09L"},
		Duration:                 4 * time.Hour, // Too long for curfew, should be adjacent
		Frequency:                24 * time.Hour,
		MinimumOperationalRunways: 1,
		CurfewStart:              &curfewStart,
		CurfewEnd:                &curfewEnd,
	}

	policy, err := NewIntelligentMaintenancePolicy(schedule)
	if err != nil {
		t.Fatalf("Failed to create policy: %v", err)
	}

	world := newMockEventWorld(simStart, simEnd, []string{"09L"})
	err = policy.GenerateEvents(context.Background(), world)
	if err != nil {
		t.Fatalf("GenerateEvents failed: %v", err)
	}

	// Verify at least one maintenance event is scheduled
	maintenanceStarts := world.CountEventsByType(event.RunwayMaintenanceStartType)
	if maintenanceStarts == 0 {
		t.Error("Expected at least one maintenance start event")
	}

	// Verify maintenance is adjacent to curfew
	foundAdjacent := false
	for _, evt := range world.events {
		if evt.Type() == event.RunwayMaintenanceStartType {
			eventTime := evt.Time()
			hour := eventTime.Hour()
			// Should be adjacent to curfew:
			// - 19:00-23:00 (maintenance ends when curfew starts)
			// - 01:00 (maintenance starts when curfew ends)
			if (hour >= 19 && hour < 23) || hour == 1 {
				foundAdjacent = true
				break
			}
		}
	}

	if !foundAdjacent {
		t.Error("Expected maintenance to be scheduled adjacent to curfew (19:00-22:59 or 01:00)")
	}
}

func TestIntelligentMaintenancePolicy_MultipleRunwaysStaggered(t *testing.T) {
	// Setup: 3 runways, should stagger maintenance
	simStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	simEnd := simStart.AddDate(0, 0, 30)

	schedule := IntelligentMaintenanceSchedule{
		RunwayDesignations:       []string{"09L", "09R", "18"},
		Duration:                 4 * time.Hour,
		Frequency:                30 * 24 * time.Hour, // Once per month
		MinimumOperationalRunways: 2, // At least 2 runways must stay operational
	}

	policy, err := NewIntelligentMaintenancePolicy(schedule)
	if err != nil {
		t.Fatalf("Failed to create policy: %v", err)
	}

	world := newMockEventWorld(simStart, simEnd, []string{"09L", "09R", "18"})
	err = policy.GenerateEvents(context.Background(), world)
	if err != nil {
		t.Fatalf("GenerateEvents failed: %v", err)
	}

	// Count maintenance events per runway
	maintenanceByRunway := make(map[string]int)
	for _, evt := range world.events {
		if evt.Type() == event.RunwayMaintenanceStartType {
			switch e := evt.(type) {
			case *event.RunwayMaintenanceStartEvent:
				maintenanceByRunway[e.RunwayID()]++
			}
		}
	}

	// Each runway should have at least one maintenance window
	for _, runway := range []string{"09L", "09R", "18"} {
		if maintenanceByRunway[runway] == 0 {
			t.Errorf("Runway %s has no scheduled maintenance", runway)
		}
	}
}

func TestIntelligentMaintenancePolicy_InvalidConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		schedule    IntelligentMaintenanceSchedule
		expectError bool
	}{
		{
			name: "invalid peak hours start",
			schedule: IntelligentMaintenanceSchedule{
				RunwayDesignations: []string{"09L"},
				Duration:           2 * time.Hour,
				Frequency:          24 * time.Hour,
				PeakHours:          &PeakHours{StartHour: 25, EndHour: 20},
			},
			expectError: true,
		},
		{
			name: "invalid peak hours end",
			schedule: IntelligentMaintenanceSchedule{
				RunwayDesignations: []string{"09L"},
				Duration:           2 * time.Hour,
				Frequency:          24 * time.Hour,
				PeakHours:          &PeakHours{StartHour: 8, EndHour: -1},
			},
			expectError: true,
		},
		{
			name: "valid configuration",
			schedule: IntelligentMaintenanceSchedule{
				RunwayDesignations: []string{"09L"},
				Duration:           2 * time.Hour,
				Frequency:          24 * time.Hour,
				PeakHours:          &PeakHours{StartHour: 8, EndHour: 20},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewIntelligentMaintenancePolicy(tt.schedule)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestIntelligentMaintenancePolicy_NonexistentRunway(t *testing.T) {
	simStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	simEnd := simStart.AddDate(0, 0, 7)

	schedule := IntelligentMaintenanceSchedule{
		RunwayDesignations:       []string{"27L"}, // Doesn't exist in mock
		Duration:                 2 * time.Hour,
		Frequency:                24 * time.Hour,
		MinimumOperationalRunways: 1,
	}

	policy, err := NewIntelligentMaintenancePolicy(schedule)
	if err != nil {
		t.Fatalf("Failed to create policy: %v", err)
	}

	world := newMockEventWorld(simStart, simEnd, []string{"09L", "09R"})
	err = policy.GenerateEvents(context.Background(), world)
	if err == nil {
		t.Error("Expected error for nonexistent runway, got nil")
	}
}
