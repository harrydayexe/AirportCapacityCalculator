package policy

import (
	"context"
	"errors"
	"math"
)

// Common errors for wind policy validation
var (
	// ErrInvalidWindSpeed indicates the wind speed is invalid
	ErrInvalidWindSpeed = errors.New("wind speed cannot be negative")

	// ErrInvalidWindDirection indicates the wind direction is invalid
	ErrInvalidWindDirection = errors.New("wind direction must be between 0 and 360 degrees")
)

// WorldState defines the interface for policies to modify world state.
// This allows the WindPolicy to set wind conditions in the simulation world.
type WorldState interface {
	SetWind(speed, direction float64) error
}

// WindPolicy models wind conditions that affect runway usability.
// Wind determines which runways can operate based on crosswind and tailwind limits.
//
// For static wind (current implementation), the wind conditions remain constant
// throughout the simulation. Future enhancements may add time-varying wind.
type WindPolicy struct {
	speedKnots     float64 // Wind speed in knots
	directionTrue  float64 // Wind direction in degrees true (0-360, where 0/360 = north, 90 = east, etc.)
}

// NewWindPolicy creates a new wind policy with validation.
// Speed is in knots, direction is in degrees true (0-360).
// Returns an error if the parameters are invalid.
func NewWindPolicy(speedKnots, directionTrue float64) (*WindPolicy, error) {
	// Validate wind speed
	if speedKnots < 0 {
		return nil, ErrInvalidWindSpeed
	}

	// Normalize direction to 0-360 range
	normalizedDirection := math.Mod(directionTrue, 360)
	if normalizedDirection < 0 {
		normalizedDirection += 360
	}

	return &WindPolicy{
		speedKnots:    speedKnots,
		directionTrue: normalizedDirection,
	}, nil
}

// Name returns the policy name.
func (p *WindPolicy) Name() string {
	return "WindPolicy"
}

// GenerateEvents sets the initial wind state in the world.
// For static wind, no events are generated - the wind remains constant.
func (p *WindPolicy) GenerateEvents(ctx context.Context, world EventWorld) error {
	// Type assert to get access to SetWind method
	worldState, ok := world.(WorldState)
	if !ok {
		return errors.New("world does not implement WorldState interface")
	}

	// Set the initial wind conditions
	if err := worldState.SetWind(p.speedKnots, p.directionTrue); err != nil {
		return err
	}

	// For static wind, no events are scheduled
	// Future: Could schedule WindChangeEvents for time-varying wind here

	return nil
}

// GetSpeed returns the wind speed in knots.
func (p *WindPolicy) GetSpeed() float64 {
	return p.speedKnots
}

// GetDirection returns the wind direction in degrees true.
func (p *WindPolicy) GetDirection() float64 {
	return p.directionTrue
}

// CalculateWindComponents computes the headwind and crosswind components for a runway.
//
// Given a runway bearing and wind direction/speed, this function decomposes the wind
// into components parallel (headwind) and perpendicular (crosswind) to the runway.
//
// Parameters:
//   - runwayBearing: Runway true bearing in degrees (0-360)
//   - windSpeed: Wind speed in knots
//   - windDirection: Wind direction in degrees true (direction wind is coming FROM)
//
// Returns:
//   - headwind: Component along runway axis (positive = headwind, negative = tailwind) in knots
//   - crosswind: Component perpendicular to runway (always positive) in knots
//
// Example:
//   Runway 09 (bearing 090°), Wind 120° at 20kt
//   Angle difference = 30°
//   Headwind = 20 * cos(30°) = 17.3kt (headwind)
//   Crosswind = 20 * |sin(30°)| = 10.0kt
func CalculateWindComponents(runwayBearing, windSpeed, windDirection float64) (headwind, crosswind float64) {
	// Calculate the angle between runway and wind direction
	// Wind direction is where wind comes FROM, so we use it directly
	angleDiff := windDirection - runwayBearing

	// Normalize angle to -180 to +180 range
	for angleDiff > 180 {
		angleDiff -= 360
	}
	for angleDiff < -180 {
		angleDiff += 360
	}

	// Convert to radians for trig functions
	angleRad := angleDiff * math.Pi / 180

	// Calculate components
	// Headwind: positive when wind is coming from ahead (helps landing, hurts takeoff acceleration)
	// Tailwind: negative headwind (hurts landing, helps takeoff acceleration)
	headwind = windSpeed * math.Cos(angleRad)

	// Crosswind: always positive (absolute value)
	crosswind = math.Abs(windSpeed * math.Sin(angleRad))

	return headwind, crosswind
}

// IsRunwayUsableInWind checks if a runway can operate under current wind conditions.
//
// A runway is unusable if:
//   - Crosswind exceeds the runway's crosswind limit
//   - Tailwind (negative headwind) exceeds the runway's tailwind limit
//
// Parameters:
//   - runwayBearing: Runway true bearing in degrees
//   - crosswindLimit: Maximum crosswind in knots (0 = no limit)
//   - tailwindLimit: Maximum tailwind in knots (0 = no limit)
//
// Returns:
//   - true if runway is usable, false otherwise
func (p *WindPolicy) IsRunwayUsableInWind(runwayBearing, crosswindLimit, tailwindLimit float64) bool {
	headwind, crosswind := CalculateWindComponents(runwayBearing, p.speedKnots, p.directionTrue)

	// Check crosswind limit (0 means no limit)
	if crosswindLimit > 0 && crosswind > crosswindLimit {
		return false
	}

	// Check tailwind limit (0 means no limit)
	// Negative headwind is tailwind
	if tailwindLimit > 0 && headwind < -tailwindLimit {
		return false
	}

	return true
}
