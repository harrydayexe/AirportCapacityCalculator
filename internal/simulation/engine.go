package simulation

import (
	"context"
	"log/slog"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/airport"
)

// Engine is the core simulation engine that calculates total movements
// based on the state modified by policies and plugins.
type Engine struct {
	logger *slog.Logger
}

// NewEngine creates a new simulation engine.
func NewEngine(logger *slog.Logger) *Engine {
	return &Engine{
		logger: logger,
	}
}

// Calculate computes the total annual movements based on the simulation state.
// It considers:
// - Available runways (after policies like maintenance are applied)
// - Operating hours (after policies like curfew are applied)
// - Minimum separation requirements from the airport configuration
func (e *Engine) Calculate(ctx context.Context, state *SimulationState) (float32, error) {
	e.logger.InfoContext(ctx, "Starting capacity calculation")

	// Determine operating seconds
	operatingSeconds := e.calculateOperatingSeconds(state)
	e.logger.InfoContext(ctx, "Operating seconds calculated", "seconds", operatingSeconds)

	// Calculate capacity based on available runways
	capacity := e.calculateRunwayCapacity(ctx, state.AvailableRunways, operatingSeconds, state.Airport)

	e.logger.InfoContext(ctx, "Total capacity calculated", "movements", capacity)

	return capacity, nil
}

// calculateOperatingSeconds determines how many seconds per year the airport operates.
func (e *Engine) calculateOperatingSeconds(state *SimulationState) float32 {
	operatingHours := state.OperatingHours
	if operatingHours == 0 {
		// Default to full year operation if not set by policies
		operatingHours = HoursPerYear
	}

	return operatingHours * SecondsPerHour
}

// calculateRunwayCapacity calculates the capacity based on runway configuration and operating time.
func (e *Engine) calculateRunwayCapacity(ctx context.Context, runways []airport.Runway, operatingSeconds float32, airport airport.Airport) float32 {
	numRunways := len(runways)

	if numRunways == 0 {
		e.logger.InfoContext(ctx, "No runways available, capacity is 0")
		return 0
	}

	separationSeconds := float32(airport.MinimumSeparation.Seconds())
	e.logger.InfoContext(ctx, "Calculating capacity",
		"numRunways", numRunways,
		"separationSeconds", separationSeconds)

	if numRunways == 1 {
		// Single runway: capacity = operating_time / separation
		capacity := operatingSeconds / separationSeconds
		e.logger.InfoContext(ctx, "Single runway capacity", "capacity", capacity)
		return capacity
	}

	if numRunways%2 == 0 {
		// Even number of runways: each runway operates in parallel
		capacity := operatingSeconds * float32(numRunways) / separationSeconds
		e.logger.InfoContext(ctx, "Even runway capacity", "capacity", capacity)
		return capacity
	}

	// Odd number of runways: calculate even + odd separately
	evenRunways := runways[:numRunways-1]
	oddRunway := runways[numRunways-1:]

	e.logger.InfoContext(ctx, "Odd number of runways, calculating separately")

	evenCapacity := e.calculateRunwayCapacity(ctx, evenRunways, operatingSeconds, airport)
	e.logger.InfoContext(ctx, "Even runway subset capacity", "capacity", evenCapacity)

	oddCapacity := e.calculateRunwayCapacity(ctx, oddRunway, operatingSeconds, airport)
	e.logger.InfoContext(ctx, "Odd runway capacity", "capacity", oddCapacity)

	totalCapacity := evenCapacity + oddCapacity
	e.logger.InfoContext(ctx, "Total capacity for odd runways", "capacity", totalCapacity)

	return totalCapacity
}
