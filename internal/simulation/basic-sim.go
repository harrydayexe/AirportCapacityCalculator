package simulation

import (
	"context"
	"log/slog"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/airport"
)

// NUM_SECONDS_IN_YEAR defines the number of seconds in a year.
const NUM_SECONDS_IN_YEAR float32 = 31536000

// BasicSim is a basic implementation of the Simulation interface where the airport exists in a perfect world.
type BasicSim struct {
	airport airport.Airport // The airport to simulate.
	logger  *slog.Logger    // The logger to use for logging.
}

// NewBasicSim creates a new BasicSim instance.
func NewBasicSim(airport airport.Airport, logger *slog.Logger) BasicSim {
	return BasicSim{
		airport: airport,
		logger:  logger,
	}
}

func (s BasicSim) Run(ctx context.Context) (float32, error) {
	// Basic assumptions:
	// A year has 365 days = 31536000 seconds
	// Runways never have to close or stop
	// The airport operates at a perfect efficiency in a perfect world.
	s.logger.InfoContext(ctx, "Starting Basic Simulation")
	return s.calculateRunwayCapacity(s.airport.Runways, ctx)
}

// calculateRunwayCapacity calculates the annual capacity of the given set of runways based on its minimum separation.
func (s BasicSim) calculateRunwayCapacity(runways []airport.Runway, ctx context.Context) (float32, error) {
	if len(runways) == 0 {
		s.logger.InfoContext(ctx, "No runways available, capacity is 0")
		return 0, nil
	} else if len(runways) == 1 {
		s.logger.InfoContext(ctx, "Single runway available, calculating capacity")
		s.logger.InfoContext(ctx, "Minimum separation between aircraft", "separation_seconds", float32(s.airport.MinimumSeparation.Seconds()))
		s.logger.InfoContext(ctx, "Number of seconds in a year", "seconds", NUM_SECONDS_IN_YEAR)

		var capacity = NUM_SECONDS_IN_YEAR / float32(s.airport.MinimumSeparation.Seconds())

		s.logger.InfoContext(ctx, "Calculated single runway capacity", "capacity", capacity)

		return capacity, nil
	} else if len(runways)%2 == 0 {
		numRunways := float32(len(runways))
		s.logger.InfoContext(ctx, "Even number of runways available, calculating capacity", "numRunways", numRunways)
		var capacity = NUM_SECONDS_IN_YEAR * numRunways / float32(s.airport.MinimumSeparation.Seconds())
		s.logger.InfoContext(ctx, "Calculated even runway capacity", "capacity", capacity)
		return capacity, nil
	} else {
		evenRunways := runways[:len(runways)-1]
		oddRunway := runways[len(runways)-1]

		s.logger.InfoContext(ctx, "Odd number of runways available, calculating capacity for even and odd runways separately")

		evenCapacity, err := s.calculateRunwayCapacity(evenRunways, ctx)
		if err != nil {
			return 0, err
		}
		s.logger.InfoContext(ctx, "Calculated even runway capacity", "capacity", evenCapacity)

		oddCapacity, err := s.calculateRunwayCapacity([]airport.Runway{oddRunway}, ctx)
		if err != nil {
			return 0, err
		}
		s.logger.InfoContext(ctx, "Calculated odd runway capacity", "capacity", oddCapacity)

		var totalCapacity = evenCapacity + oddCapacity
		s.logger.InfoContext(ctx, "Calculated total runway capacity for odd number of runways", "capacity", totalCapacity)
		return totalCapacity, nil
	}
}
