package policy

import (
	"context"
	"testing"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/event"
)

func TestNewCurfewPolicy(t *testing.T) {
	startTime := time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 2, 6, 0, 0, 0, time.UTC)

	policy, err := NewCurfewPolicy(startTime, endTime)
	if err != nil {
		t.Fatalf("NewCurfewPolicy returned error: %v", err)
	}
	if policy == nil {
		t.Fatal("NewCurfewPolicy returned nil")
	}
	if !policy.startTime.Equal(startTime) {
		t.Errorf("expected startTime %v, got %v", startTime, policy.startTime)
	}
	if !policy.endTime.Equal(endTime) {
		t.Errorf("expected endTime %v, got %v", endTime, policy.endTime)
	}
}

func TestCurfewPolicy_Name(t *testing.T) {
	startTime := time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 2, 6, 0, 0, 0, time.UTC)

	policy, err := NewCurfewPolicy(startTime, endTime)
	if err != nil {
		t.Fatalf("NewCurfewPolicy returned error: %v", err)
	}
	name := policy.Name()
	expectedName := "CurfewPolicy"
	if name != expectedName {
		t.Errorf("expected name %q, got %q", expectedName, name)
	}
}

func TestCurfewPolicy_GenerateEvents(t *testing.T) {
	tests := []struct {
		name                    string
		curfewStartTime         time.Time
		curfewEndTime           time.Time
		simStartTime            time.Time
		simEndTime              time.Time
		expectedCurfewStarts    int
		expectedCurfewEnds      int
		verifyFirstEventTime    bool
		expectedFirstEventHour  int
		expectedFirstEventMin   int
	}{
		{
			name:                    "7 hour nightly curfew (11pm-6am) for 1 week",
			curfewStartTime:         time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC),
			curfewEndTime:           time.Date(2024, 1, 2, 6, 0, 0, 0, time.UTC),
			simStartTime:            time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			simEndTime:              time.Date(2024, 1, 8, 6, 0, 0, 0, time.UTC), // Extended to include last curfew end
			expectedCurfewStarts:    7,
			expectedCurfewEnds:      7,
			verifyFirstEventTime:    true,
			expectedFirstEventHour:  23,
			expectedFirstEventMin:   0,
		},
		{
			name:                    "Full year simulation",
			curfewStartTime:         time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC),
			curfewEndTime:           time.Date(2024, 1, 2, 6, 0, 0, 0, time.UTC),
			simStartTime:            time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			simEndTime:              time.Date(2025, 1, 1, 6, 0, 0, 0, time.UTC), // Extended to include last curfew end
			expectedCurfewStarts:    366, // 2024 is a leap year
			expectedCurfewEnds:      366,
			verifyFirstEventTime:    false,
		},
		{
			name:                    "4 hour curfew (midnight-4am)",
			curfewStartTime:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			curfewEndTime:           time.Date(2024, 1, 1, 4, 0, 0, 0, time.UTC),
			simStartTime:            time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			simEndTime:              time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC), // 3 days
			expectedCurfewStarts:    3,
			expectedCurfewEnds:      3,
			verifyFirstEventTime:    true,
			expectedFirstEventHour:  0,
			expectedFirstEventMin:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := NewCurfewPolicy(tt.curfewStartTime, tt.curfewEndTime)
			if err != nil {
				t.Fatalf("NewCurfewPolicy failed: %v", err)
			}

			world := newMockEventWorld(tt.simStartTime, tt.simEndTime, []string{"09L"})

			err = policy.GenerateEvents(context.Background(), world)
			if err != nil {
				t.Fatalf("GenerateEvents failed: %v", err)
			}

			// Count events by type
			curfewStarts := world.CountEventsByType(event.CurfewStartType)
			curfewEnds := world.CountEventsByType(event.CurfewEndType)

			if curfewStarts != tt.expectedCurfewStarts {
				t.Errorf("expected %d curfew start events, got %d", tt.expectedCurfewStarts, curfewStarts)
			}

			if curfewEnds != tt.expectedCurfewEnds {
				t.Errorf("expected %d curfew end events, got %d", tt.expectedCurfewEnds, curfewEnds)
			}

			// Verify first event timing
			if tt.verifyFirstEventTime && len(world.GetEvents()) > 0 {
				firstEvent := world.GetEvents()[0]
				if firstEvent.Type() == event.CurfewStartType {
					eventTime := firstEvent.Time()
					if eventTime.Hour() != tt.expectedFirstEventHour || eventTime.Minute() != tt.expectedFirstEventMin {
						t.Errorf("expected first curfew start at %02d:%02d, got %02d:%02d",
							tt.expectedFirstEventHour, tt.expectedFirstEventMin,
							eventTime.Hour(), eventTime.Minute())
					}
				}
			}

			// Verify events are in chronological pairs (start, then end)
			events := world.GetEvents()
			for i := 0; i < len(events)-1; i += 2 {
				if i+1 >= len(events) {
					break
				}
				startEvent := events[i]
				endEvent := events[i+1]

				if startEvent.Type() != event.CurfewStartType {
					t.Errorf("event %d should be CurfewStart, got %s", i, startEvent.Type())
				}
				if endEvent.Type() != event.CurfewEndType {
					t.Errorf("event %d should be CurfewEnd, got %s", i+1, endEvent.Type())
				}

				// End should be after start
				if !endEvent.Time().After(startEvent.Time()) {
					t.Errorf("curfew end (%v) should be after curfew start (%v)",
						endEvent.Time(), startEvent.Time())
				}
			}
		})
	}
}

func TestCurfewPolicy_GenerateEvents_OvernightCurfew(t *testing.T) {
	// Test an overnight curfew (11pm-6am)
	curfewStartTime := time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC)
	curfewEndTime := time.Date(2024, 1, 2, 6, 0, 0, 0, time.UTC)

	policy, err := NewCurfewPolicy(curfewStartTime, curfewEndTime)
	if err != nil {
		t.Fatalf("NewCurfewPolicy failed: %v", err)
	}

	simStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	simEnd := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC) // 2 days

	world := newMockEventWorld(simStart, simEnd, []string{"09L"})

	err = policy.GenerateEvents(context.Background(), world)
	if err != nil {
		t.Fatalf("GenerateEvents failed: %v", err)
	}

	// Verify the curfew end times are on the next day
	events := world.GetEvents()
	for i := 0; i < len(events)-1; i += 2 {
		if i+1 >= len(events) {
			break
		}
		startEvent := events[i]
		endEvent := events[i+1]

		// For overnight curfew, end should be on the next day
		if endEvent.Time().Day() != startEvent.Time().Day()+1 {
			t.Errorf("overnight curfew end should be next day: start=%v, end=%v",
				startEvent.Time(), endEvent.Time())
		}
	}
}

func TestNewCurfewPolicy_Validation(t *testing.T) {
	tests := []struct {
		name        string
		startTime   time.Time
		endTime     time.Time
		expectError error
	}{
		{
			name:        "Valid curfew",
			startTime:   time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC),
			endTime:     time.Date(2024, 1, 2, 6, 0, 0, 0, time.UTC),
			expectError: nil,
		},
		{
			name:        "End time before start time",
			startTime:   time.Date(2024, 1, 2, 6, 0, 0, 0, time.UTC),
			endTime:     time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC),
			expectError: ErrInvalidCurfewTime,
		},
		{
			name:        "End time equal to start time",
			startTime:   time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC),
			endTime:     time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC),
			expectError: ErrInvalidCurfewTime,
		},
		{
			name:        "Curfew exceeds maximum duration",
			startTime:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			endTime:     time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC), // 45 days
			expectError: ErrCurfewTooLong,
		},
		{
			name:        "Curfew at maximum duration boundary",
			startTime:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			endTime:     time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC), // 30 days exactly
			expectError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := NewCurfewPolicy(tt.startTime, tt.endTime)

			if tt.expectError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectError)
				} else if err != tt.expectError {
					t.Errorf("expected error %v, got %v", tt.expectError, err)
				}
				if policy != nil {
					t.Error("expected nil policy when error is returned")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if policy == nil {
					t.Error("expected valid policy, got nil")
				}
			}
		})
	}
}
