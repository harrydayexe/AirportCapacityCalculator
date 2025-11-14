package policy

import (
	"fmt"
	"math"
	"time"
)

// DiurnalWindPattern generates a realistic daily wind pattern with morning calm,
// afternoon build-up, and evening decrease. This models typical land-sea breeze
// or thermal wind patterns.
//
// Parameters:
//   - startDate: The date to start the pattern (time will be set to midnight)
//   - days: Number of days to generate the pattern for
//   - morningSpeed: Wind speed at 06:00 (knots)
//   - afternoonSpeed: Peak wind speed at 15:00 (knots)
//   - eveningSpeed: Wind speed at 21:00 (knots)
//   - direction: Predominant wind direction in degrees true (constant throughout day)
//
// Returns a wind schedule with 4 changes per day: midnight calm, morning, afternoon peak, evening.
func DiurnalWindPattern(startDate time.Time, days int, morningSpeed, afternoonSpeed, eveningSpeed, direction float64) []WindChange {
	// Normalize start to midnight
	start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())

	schedule := make([]WindChange, 0, days*4)

	for day := 0; day < days; day++ {
		currentDay := start.AddDate(0, 0, day)

		// Midnight: Calm
		schedule = append(schedule, WindChange{
			Timestamp:     currentDay,
			SpeedKnots:    0,
			DirectionTrue: direction,
		})

		// 06:00: Morning light winds
		schedule = append(schedule, WindChange{
			Timestamp:     currentDay.Add(6 * time.Hour),
			SpeedKnots:    morningSpeed,
			DirectionTrue: direction,
		})

		// 15:00: Afternoon peak
		schedule = append(schedule, WindChange{
			Timestamp:     currentDay.Add(15 * time.Hour),
			SpeedKnots:    afternoonSpeed,
			DirectionTrue: direction,
		})

		// 21:00: Evening decrease
		schedule = append(schedule, WindChange{
			Timestamp:     currentDay.Add(21 * time.Hour),
			SpeedKnots:    eveningSpeed,
			DirectionTrue: direction,
		})
	}

	return schedule
}

// ConstantWindPattern generates a simple constant wind schedule with a single wind change.
// Useful for testing or modeling steady-state conditions.
//
// Parameters:
//   - timestamp: When the wind condition takes effect
//   - speedKnots: Wind speed in knots
//   - directionTrue: Wind direction in degrees true
func ConstantWindPattern(timestamp time.Time, speedKnots, directionTrue float64) []WindChange {
	return []WindChange{
		{
			Timestamp:     timestamp,
			SpeedKnots:    speedKnots,
			DirectionTrue: directionTrue,
		},
	}
}

// FrontalPassagePattern models an abrupt wind shift typical of a cold front passage.
// Wind shifts from pre-frontal to post-frontal conditions at a specified time.
//
// Parameters:
//   - passageTime: When the front passes
//   - preFrontalSpeed: Wind speed before the front (knots)
//   - preFrontalDirection: Wind direction before the front (degrees true)
//   - postFrontalSpeed: Wind speed after the front (knots)
//   - postFrontalDirection: Wind direction after the front (degrees true)
//
// Returns a schedule with two wind changes: one at passageTime-1h (pre-frontal) and one at passageTime (post-frontal).
func FrontalPassagePattern(passageTime time.Time, preFrontalSpeed, preFrontalDirection, postFrontalSpeed, postFrontalDirection float64) []WindChange {
	return []WindChange{
		{
			Timestamp:     passageTime.Add(-1 * time.Hour),
			SpeedKnots:    preFrontalSpeed,
			DirectionTrue: preFrontalDirection,
		},
		{
			Timestamp:     passageTime,
			SpeedKnots:    postFrontalSpeed,
			DirectionTrue: postFrontalDirection,
		},
	}
}

// LinearWindTransition generates a smooth wind transition from initial to final conditions
// over a specified duration. The wind speed and direction change linearly over time.
//
// Parameters:
//   - startTime: When the transition begins
//   - duration: How long the transition takes
//   - steps: Number of intermediate steps (more steps = smoother transition)
//   - initialSpeed: Starting wind speed (knots)
//   - initialDirection: Starting wind direction (degrees true)
//   - finalSpeed: Ending wind speed (knots)
//   - finalDirection: Ending wind direction (degrees true)
//
// Note: Direction transitions always take the shortest angular path (e.g., 350° to 10° goes through 360°, not backwards through 180°).
func LinearWindTransition(startTime time.Time, duration time.Duration, steps int, initialSpeed, initialDirection, finalSpeed, finalDirection float64) ([]WindChange, error) {
	if steps < 2 {
		return nil, fmt.Errorf("steps must be at least 2, got %d", steps)
	}

	schedule := make([]WindChange, steps)

	// Calculate shortest angular path for direction change
	directionDelta := finalDirection - initialDirection
	if directionDelta > 180 {
		directionDelta -= 360
	} else if directionDelta < -180 {
		directionDelta += 360
	}

	speedDelta := finalSpeed - initialSpeed
	stepDuration := duration / time.Duration(steps-1)

	for i := 0; i < steps; i++ {
		progress := float64(i) / float64(steps-1)

		speed := initialSpeed + (speedDelta * progress)
		direction := initialDirection + (directionDelta * progress)

		// Normalize direction to 0-360
		direction = math.Mod(direction, 360)
		if direction < 0 {
			direction += 360
		}

		schedule[i] = WindChange{
			Timestamp:     startTime.Add(time.Duration(i) * stepDuration),
			SpeedKnots:    speed,
			DirectionTrue: direction,
		}
	}

	return schedule, nil
}

// SeasonalWindPattern generates a wind pattern that varies by season throughout the year.
// Useful for modeling prevailing winds that shift with the seasons.
//
// Parameters:
//   - year: The year to generate the pattern for
//   - location: Timezone location for the pattern
//   - winterSpeed, springSpeed, summerSpeed, fallSpeed: Average wind speeds per season (knots)
//   - winterDir, springDir, summerDir, fallDir: Predominant wind directions per season (degrees true)
//
// Returns a schedule with wind changes at the start of each season.
func SeasonalWindPattern(year int, location *time.Location, winterSpeed, springSpeed, summerSpeed, fallSpeed, winterDir, springDir, summerDir, fallDir float64) []WindChange {
	return []WindChange{
		// Winter (January 1)
		{
			Timestamp:     time.Date(year, 1, 1, 0, 0, 0, 0, location),
			SpeedKnots:    winterSpeed,
			DirectionTrue: winterDir,
		},
		// Spring (March 20 - approximate equinox)
		{
			Timestamp:     time.Date(year, 3, 20, 0, 0, 0, 0, location),
			SpeedKnots:    springSpeed,
			DirectionTrue: springDir,
		},
		// Summer (June 21 - approximate solstice)
		{
			Timestamp:     time.Date(year, 6, 21, 0, 0, 0, 0, location),
			SpeedKnots:    summerSpeed,
			DirectionTrue: summerDir,
		},
		// Fall (September 22 - approximate equinox)
		{
			Timestamp:     time.Date(year, 9, 22, 0, 0, 0, 0, location),
			SpeedKnots:    fallSpeed,
			DirectionTrue: fallDir,
		},
	}
}

// CombineWindSchedules merges multiple wind schedules into a single schedule and sorts them chronologically.
// This is useful for combining different wind patterns (e.g., seasonal + diurnal).
//
// Note: If multiple wind changes occur at the exact same timestamp, the last one in the input order takes precedence.
func CombineWindSchedules(schedules ...[]WindChange) []WindChange {
	totalSize := 0
	for _, schedule := range schedules {
		totalSize += len(schedule)
	}

	combined := make([]WindChange, 0, totalSize)
	for _, schedule := range schedules {
		combined = append(combined, schedule...)
	}

	SortSchedule(combined)
	return combined
}
