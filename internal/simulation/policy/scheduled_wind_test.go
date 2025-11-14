package policy

import (
	"context"
	"testing"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/event"
)

// TestNewScheduledWindPolicy tests the constructor
func TestNewScheduledWindPolicy(t *testing.T) {
	tests := []struct {
		name        string
		schedule    []WindChange
		expectError bool
		errorType   error
	}{
		{
			name: "valid single change",
			schedule: []WindChange{
				{time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), 15, 270},
			},
			expectError: false,
		},
		{
			name: "valid multiple changes",
			schedule: []WindChange{
				{time.Date(2024, 1, 1, 6, 0, 0, 0, time.UTC), 5, 90},
				{time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), 15, 270},
				{time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC), 20, 270},
			},
			expectError: false,
		},
		{
			name:        "empty schedule",
			schedule:    []WindChange{},
			expectError: true,
			errorType:   ErrEmptyWindSchedule,
		},
		{
			name: "negative wind speed",
			schedule: []WindChange{
				{time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), -5, 270},
			},
			expectError: true,
			errorType:   ErrInvalidWindSpeed,
		},
		{
			name: "not chronological",
			schedule: []WindChange{
				{time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC), 20, 270},
				{time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), 15, 270},
			},
			expectError: true,
			errorType:   ErrWindScheduleNotChronological,
		},
		{
			name: "direction normalization",
			schedule: []WindChange{
				{time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), 15, 450}, // Should normalize to 90
			},
			expectError: false,
		},
		{
			name: "negative direction normalization",
			schedule: []WindChange{
				{time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), 15, -90}, // Should normalize to 270
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := NewScheduledWindPolicy(tt.schedule)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if policy == nil {
				t.Error("Expected non-nil policy")
			}
		})
	}
}

// TestScheduledWindPolicyName tests the Name method
func TestScheduledWindPolicyName(t *testing.T) {
	policy, _ := NewScheduledWindPolicy([]WindChange{
		{time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), 15, 270},
	})

	if policy.Name() != "ScheduledWindPolicy" {
		t.Errorf("Expected name 'ScheduledWindPolicy', got '%s'", policy.Name())
	}
}

// TestScheduledWindPolicyGenerateEvents tests event generation
func TestScheduledWindPolicyGenerateEvents(t *testing.T) {
	simStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	simEnd := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		schedule      []WindChange
		expectedCount int
	}{
		{
			name: "all events within period",
			schedule: []WindChange{
				{time.Date(2024, 1, 1, 6, 0, 0, 0, time.UTC), 5, 90},
				{time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), 15, 270},
				{time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC), 20, 270},
			},
			expectedCount: 3,
		},
		{
			name: "some events outside period",
			schedule: []WindChange{
				{time.Date(2023, 12, 31, 23, 0, 0, 0, time.UTC), 5, 90},  // Before
				{time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), 15, 270},  // Within
				{time.Date(2024, 1, 3, 1, 0, 0, 0, time.UTC), 20, 270},   // After
			},
			expectedCount: 1,
		},
		{
			name: "all events outside period",
			schedule: []WindChange{
				{time.Date(2023, 12, 31, 12, 0, 0, 0, time.UTC), 5, 90},
				{time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC), 15, 270},
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := NewScheduledWindPolicy(tt.schedule)
			if err != nil {
				t.Fatalf("Failed to create policy: %v", err)
			}

			mockWorld := newMockEventWorld(simStart, simEnd, nil)

			err = policy.GenerateEvents(context.Background(), mockWorld)
			if err != nil {
				t.Fatalf("GenerateEvents failed: %v", err)
			}

			events := mockWorld.GetEvents()
			if len(events) != tt.expectedCount {
				t.Errorf("Expected %d events, got %d", tt.expectedCount, len(events))
			}

			// Verify all scheduled events are WindChangeEvents
			for _, evt := range events {
				if evt.Type() != event.WindChangeType {
					t.Errorf("Expected WindChangeType, got %v", evt.Type())
				}
			}
		})
	}
}

// TestScheduledWindPolicyGetSchedule tests the GetSchedule method
func TestScheduledWindPolicyGetSchedule(t *testing.T) {
	original := []WindChange{
		{time.Date(2024, 1, 1, 6, 0, 0, 0, time.UTC), 5, 90},
		{time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), 15, 270},
	}

	policy, err := NewScheduledWindPolicy(original)
	if err != nil {
		t.Fatalf("Failed to create policy: %v", err)
	}

	returned := policy.GetSchedule()

	if len(returned) != len(original) {
		t.Errorf("Expected %d wind changes, got %d", len(original), len(returned))
	}

	// Verify it's a copy, not the same slice
	returned[0].SpeedKnots = 999
	actual := policy.GetSchedule()
	if actual[0].SpeedKnots == 999 {
		t.Error("GetSchedule should return a copy, not the original slice")
	}
}

// TestScheduledWindPolicyGetWindAt tests the GetWindAt method
func TestScheduledWindPolicyGetWindAt(t *testing.T) {
	schedule := []WindChange{
		{time.Date(2024, 1, 1, 6, 0, 0, 0, time.UTC), 5, 90},
		{time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), 15, 270},
		{time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC), 25, 270},
	}

	policy, err := NewScheduledWindPolicy(schedule)
	if err != nil {
		t.Fatalf("Failed to create policy: %v", err)
	}

	tests := []struct {
		name              string
		queryTime         time.Time
		expectedSpeed     float64
		expectedDirection float64
	}{
		{
			name:              "before all changes (calm)",
			queryTime:         time.Date(2024, 1, 1, 3, 0, 0, 0, time.UTC),
			expectedSpeed:     0,
			expectedDirection: 0,
		},
		{
			name:              "at first change",
			queryTime:         time.Date(2024, 1, 1, 6, 0, 0, 0, time.UTC),
			expectedSpeed:     5,
			expectedDirection: 90,
		},
		{
			name:              "between first and second",
			queryTime:         time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
			expectedSpeed:     5,
			expectedDirection: 90,
		},
		{
			name:              "at second change",
			queryTime:         time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			expectedSpeed:     15,
			expectedDirection: 270,
		},
		{
			name:              "after all changes",
			queryTime:         time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC),
			expectedSpeed:     25,
			expectedDirection: 270,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			speed, direction := policy.GetWindAt(tt.queryTime)

			if speed != tt.expectedSpeed {
				t.Errorf("Expected speed %f, got %f", tt.expectedSpeed, speed)
			}

			if direction != tt.expectedDirection {
				t.Errorf("Expected direction %f, got %f", tt.expectedDirection, direction)
			}
		})
	}
}

// TestSortSchedule tests the sort utility function
func TestSortSchedule(t *testing.T) {
	schedule := []WindChange{
		{time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC), 25, 270},
		{time.Date(2024, 1, 1, 6, 0, 0, 0, time.UTC), 5, 90},
		{time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), 15, 270},
	}

	SortSchedule(schedule)

	// Verify chronological order
	for i := 1; i < len(schedule); i++ {
		if !schedule[i].Timestamp.After(schedule[i-1].Timestamp) {
			t.Errorf("Schedule not sorted: entry %d (%v) not after entry %d (%v)",
				i, schedule[i].Timestamp, i-1, schedule[i-1].Timestamp)
		}
	}

	// Verify specific order
	if schedule[0].SpeedKnots != 5 || schedule[1].SpeedKnots != 15 || schedule[2].SpeedKnots != 25 {
		t.Error("Schedule not sorted correctly by timestamp")
	}
}
