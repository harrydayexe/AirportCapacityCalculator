package policy

import (
	"testing"
	"time"
)

// TestDiurnalWindPattern tests the diurnal wind pattern generator
func TestDiurnalWindPattern(t *testing.T) {
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	pattern := DiurnalWindPattern(startDate, 2, 5, 20, 10, 270)

	// Should have 4 changes per day × 2 days = 8 changes
	if len(pattern) != 8 {
		t.Errorf("Expected 8 wind changes, got %d", len(pattern))
	}

	// Check day 1 schedule
	expectedDay1 := []struct {
		hour  int
		speed float64
	}{
		{0, 0},   // Midnight: calm
		{6, 5},   // Morning
		{15, 20}, // Afternoon peak
		{21, 10}, // Evening
	}

	for i, expected := range expectedDay1 {
		change := pattern[i]

		if change.Timestamp.Hour() != expected.hour {
			t.Errorf("Day 1, change %d: expected hour %d, got %d",
				i, expected.hour, change.Timestamp.Hour())
		}

		if change.SpeedKnots != expected.speed {
			t.Errorf("Day 1, change %d: expected speed %f, got %f",
				i, expected.speed, change.SpeedKnots)
		}

		if change.DirectionTrue != 270 {
			t.Errorf("Day 1, change %d: expected direction 270, got %f",
				i, change.DirectionTrue)
		}
	}

	// Check day 2 starts at correct time
	if !pattern[4].Timestamp.Equal(startDate.AddDate(0, 0, 1)) {
		t.Errorf("Day 2 should start at %v, got %v",
			startDate.AddDate(0, 0, 1), pattern[4].Timestamp)
	}
}

// TestDiurnalWindPatternWithNonMidnightStart tests that diurnal pattern normalizes to midnight
func TestDiurnalWindPatternWithNonMidnightStart(t *testing.T) {
	// Start at 14:30
	startDate := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)
	pattern := DiurnalWindPattern(startDate, 1, 5, 20, 10, 90)

	// First change should be at midnight, not 14:30
	expected := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if !pattern[0].Timestamp.Equal(expected) {
		t.Errorf("Expected first change at midnight %v, got %v", expected, pattern[0].Timestamp)
	}
}

// TestConstantWindPattern tests the constant wind pattern generator
func TestConstantWindPattern(t *testing.T) {
	timestamp := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	pattern := ConstantWindPattern(timestamp, 15, 180)

	if len(pattern) != 1 {
		t.Errorf("Expected 1 wind change, got %d", len(pattern))
	}

	if !pattern[0].Timestamp.Equal(timestamp) {
		t.Errorf("Expected timestamp %v, got %v", timestamp, pattern[0].Timestamp)
	}

	if pattern[0].SpeedKnots != 15 {
		t.Errorf("Expected speed 15, got %f", pattern[0].SpeedKnots)
	}

	if pattern[0].DirectionTrue != 180 {
		t.Errorf("Expected direction 180, got %f", pattern[0].DirectionTrue)
	}
}

// TestFrontalPassagePattern tests the frontal passage pattern generator
func TestFrontalPassagePattern(t *testing.T) {
	passageTime := time.Date(2024, 3, 15, 18, 0, 0, 0, time.UTC)
	pattern := FrontalPassagePattern(passageTime, 10, 180, 25, 270)

	if len(pattern) != 2 {
		t.Errorf("Expected 2 wind changes, got %d", len(pattern))
	}

	// Pre-frontal (1 hour before)
	expectedPreTime := passageTime.Add(-1 * time.Hour)
	if !pattern[0].Timestamp.Equal(expectedPreTime) {
		t.Errorf("Pre-frontal time: expected %v, got %v", expectedPreTime, pattern[0].Timestamp)
	}
	if pattern[0].SpeedKnots != 10 {
		t.Errorf("Pre-frontal speed: expected 10, got %f", pattern[0].SpeedKnots)
	}
	if pattern[0].DirectionTrue != 180 {
		t.Errorf("Pre-frontal direction: expected 180, got %f", pattern[0].DirectionTrue)
	}

	// Post-frontal (at passage time)
	if !pattern[1].Timestamp.Equal(passageTime) {
		t.Errorf("Post-frontal time: expected %v, got %v", passageTime, pattern[1].Timestamp)
	}
	if pattern[1].SpeedKnots != 25 {
		t.Errorf("Post-frontal speed: expected 25, got %f", pattern[1].SpeedKnots)
	}
	if pattern[1].DirectionTrue != 270 {
		t.Errorf("Post-frontal direction: expected 270, got %f", pattern[1].DirectionTrue)
	}
}

// TestLinearWindTransition tests the linear wind transition generator
func TestLinearWindTransition(t *testing.T) {
	startTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	duration := 4 * time.Hour
	steps := 5

	pattern, err := LinearWindTransition(startTime, duration, steps, 10, 90, 30, 180)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(pattern) != steps {
		t.Errorf("Expected %d steps, got %d", steps, len(pattern))
	}

	// Check first step (initial conditions)
	if pattern[0].SpeedKnots != 10 {
		t.Errorf("Initial speed: expected 10, got %f", pattern[0].SpeedKnots)
	}
	if pattern[0].DirectionTrue != 90 {
		t.Errorf("Initial direction: expected 90, got %f", pattern[0].DirectionTrue)
	}

	// Check last step (final conditions)
	if pattern[4].SpeedKnots != 30 {
		t.Errorf("Final speed: expected 30, got %f", pattern[4].SpeedKnots)
	}
	if pattern[4].DirectionTrue != 180 {
		t.Errorf("Final direction: expected 180, got %f", pattern[4].DirectionTrue)
	}

	// Check middle step (should be halfway)
	if pattern[2].SpeedKnots != 20 {
		t.Errorf("Middle speed: expected 20, got %f", pattern[2].SpeedKnots)
	}
	if pattern[2].DirectionTrue != 135 {
		t.Errorf("Middle direction: expected 135, got %f", pattern[2].DirectionTrue)
	}

	// Check timestamps are evenly spaced
	stepDuration := duration / time.Duration(steps-1)
	for i := 0; i < steps; i++ {
		expectedTime := startTime.Add(time.Duration(i) * stepDuration)
		if !pattern[i].Timestamp.Equal(expectedTime) {
			t.Errorf("Step %d time: expected %v, got %v", i, expectedTime, pattern[i].Timestamp)
		}
	}
}

// TestLinearWindTransitionShortestPath tests that direction changes take the shortest angular path
func TestLinearWindTransitionShortestPath(t *testing.T) {
	startTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name             string
		initialDirection float64
		finalDirection   float64
		expectedMiddle   float64 // Direction at middle step
	}{
		{
			name:             "350 to 10 (should go through 360, not backwards)",
			initialDirection: 350,
			finalDirection:   10,
			expectedMiddle:   0, // Halfway between 350 and 10 going through 360
		},
		{
			name:             "10 to 350 (should go backwards through 360)",
			initialDirection: 10,
			finalDirection:   350,
			expectedMiddle:   0, // Halfway between 10 and 350 going through 360
		},
		{
			name:             "90 to 180 (normal forward)",
			initialDirection: 90,
			finalDirection:   180,
			expectedMiddle:   135,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern, err := LinearWindTransition(startTime, 2*time.Hour, 3, 10, tt.initialDirection, 10, tt.finalDirection)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			middle := pattern[1].DirectionTrue
			if middle != tt.expectedMiddle {
				t.Errorf("Expected middle direction %f, got %f", tt.expectedMiddle, middle)
			}
		})
	}
}

// TestLinearWindTransitionInvalidSteps tests error handling for invalid step counts
func TestLinearWindTransitionInvalidSteps(t *testing.T) {
	startTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		steps int
	}{
		{"zero steps", 0},
		{"one step", 1},
		{"negative steps", -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LinearWindTransition(startTime, 1*time.Hour, tt.steps, 10, 90, 20, 180)
			if err == nil {
				t.Error("Expected error for invalid steps, got nil")
			}
		})
	}
}

// TestSeasonalWindPattern tests the seasonal wind pattern generator
func TestSeasonalWindPattern(t *testing.T) {
	pattern := SeasonalWindPattern(2024, time.UTC, 15, 10, 5, 12, 270, 180, 90, 225)

	if len(pattern) != 4 {
		t.Errorf("Expected 4 seasonal changes, got %d", len(pattern))
	}

	// Check each season
	seasons := []struct {
		month     time.Month
		day       int
		speed     float64
		direction float64
	}{
		{time.January, 1, 15, 270},   // Winter
		{time.March, 20, 10, 180},    // Spring
		{time.June, 21, 5, 90},       // Summer
		{time.September, 22, 12, 225}, // Fall
	}

	for i, season := range seasons {
		change := pattern[i]

		if change.Timestamp.Month() != season.month {
			t.Errorf("Season %d: expected month %v, got %v", i, season.month, change.Timestamp.Month())
		}

		if change.Timestamp.Day() != season.day {
			t.Errorf("Season %d: expected day %d, got %d", i, season.day, change.Timestamp.Day())
		}

		if change.SpeedKnots != season.speed {
			t.Errorf("Season %d: expected speed %f, got %f", i, season.speed, change.SpeedKnots)
		}

		if change.DirectionTrue != season.direction {
			t.Errorf("Season %d: expected direction %f, got %f", i, season.direction, change.DirectionTrue)
		}
	}
}

// TestCombineWindSchedules tests combining multiple wind schedules
func TestCombineWindSchedules(t *testing.T) {
	schedule1 := []WindChange{
		{time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), 10, 90},
		{time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC), 15, 180},
	}

	schedule2 := []WindChange{
		{time.Date(2024, 1, 1, 6, 0, 0, 0, time.UTC), 5, 270},
		{time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 20, 270},
	}

	combined := CombineWindSchedules(schedule1, schedule2)

	// Should have all 4 changes
	if len(combined) != 4 {
		t.Errorf("Expected 4 combined changes, got %d", len(combined))
	}

	// Should be sorted chronologically
	expectedHours := []int{6, 12, 15, 18}
	for i, expectedHour := range expectedHours {
		if combined[i].Timestamp.Hour() != expectedHour {
			t.Errorf("Change %d: expected hour %d, got %d",
				i, expectedHour, combined[i].Timestamp.Hour())
		}
	}
}

// TestCombineWindSchedulesEmpty tests combining with empty schedules
func TestCombineWindSchedulesEmpty(t *testing.T) {
	schedule1 := []WindChange{
		{time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), 10, 90},
	}

	combined := CombineWindSchedules(schedule1, []WindChange{}, nil)

	if len(combined) != 1 {
		t.Errorf("Expected 1 change, got %d", len(combined))
	}
}
