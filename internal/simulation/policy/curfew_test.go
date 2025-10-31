package policy

import (
	"context"
	"testing"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/airport"
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

func TestCurfewPolicy_Apply(t *testing.T) {
	tests := []struct {
		name               string
		startTime          time.Time
		endTime            time.Time
		initialHours       float32
		expectedReduction  float32 // daily reduction hours * 365
		expectOperatingSet bool
	}{
		{
			name:               "7 hour nightly curfew (11pm-6am)",
			startTime:          time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC),
			endTime:            time.Date(2024, 1, 2, 6, 0, 0, 0, time.UTC),
			initialHours:       8760, // Full year
			expectedReduction:  7 * 365,
			expectOperatingSet: true,
		},
		{
			name:               "4 hour curfew (midnight-4am)",
			startTime:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			endTime:            time.Date(2024, 1, 1, 4, 0, 0, 0, time.UTC),
			initialHours:       8760,
			expectedReduction:  4 * 365,
			expectOperatingSet: true,
		},
		{
			name:               "12 hour curfew (8pm-8am)",
			startTime:          time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC),
			endTime:            time.Date(2024, 1, 2, 8, 0, 0, 0, time.UTC),
			initialHours:       8760,
			expectedReduction:  12 * 365,
			expectOperatingSet: true,
		},
		{
			name:               "Zero initial hours defaults to full year",
			startTime:          time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC),
			endTime:            time.Date(2024, 1, 2, 6, 0, 0, 0, time.UTC),
			initialHours:       0, // Should default to 8760
			expectedReduction:  7 * 365,
			expectOperatingSet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := NewCurfewPolicy(tt.startTime, tt.endTime)
			if err != nil {
				t.Fatalf("NewCurfewPolicy failed: %v", err)
			}
			state := &mockSimulationState{
				operatingHours: tt.initialHours,
				availableRunways: []airport.Runway{
					{RunwayDesignation: "09L"},
				},
			}

			err = policy.Apply(context.Background(), state, testLogger())
			if err != nil {
				t.Fatalf("Apply failed: %v", err)
			}

			if tt.expectOperatingSet {
				expectedHours := tt.initialHours
				if expectedHours == 0 {
					expectedHours = 8760 // default
				}
				expectedHours -= tt.expectedReduction

				// Allow for small floating point differences
				diff := state.operatingHours - expectedHours
				if diff < -0.01 || diff > 0.01 {
					t.Errorf("expected operating hours %v, got %v (reduction: %v)", expectedHours, state.operatingHours, tt.expectedReduction)
				}

				// Ensure hours never go negative
				if state.operatingHours < 0 {
					t.Errorf("operating hours should never be negative, got %v", state.operatingHours)
				}
			}
		})
	}
}

func TestCurfewPolicy_Apply_InvalidState(t *testing.T) {
	startTime := time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 2, 6, 0, 0, 0, time.UTC)
	policy, err := NewCurfewPolicy(startTime, endTime)
	if err != nil {
		t.Fatalf("NewCurfewPolicy failed: %v", err)
	}

	// Test with an invalid state type
	err = policy.Apply(context.Background(), "invalid state", testLogger())
	if err == nil {
		t.Error("expected error for invalid state type, got nil")
	}
}

func TestCurfewPolicy_Apply_MultiDayCurfew(t *testing.T) {
	// Test a curfew that spans multiple days (should normalize to daily)
	startTime := time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 3, 6, 0, 0, 0, time.UTC) // 31 hours later

	policy, err := NewCurfewPolicy(startTime, endTime)
	if err != nil {
		t.Fatalf("NewCurfewPolicy failed: %v", err)
	}
	state := &mockSimulationState{
		operatingHours: 8760,
	}

	err = policy.Apply(context.Background(), state, testLogger())
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Should normalize to daily curfew (31 hours % 24 = 7 hours)
	expectedReduction := float32(7 * 365)
	expectedHours := float32(8760) - expectedReduction

	diff := state.operatingHours - expectedHours
	if diff < -0.01 || diff > 0.01 {
		t.Errorf("multi-day curfew should normalize: expected %v hours, got %v", expectedHours, state.operatingHours)
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
