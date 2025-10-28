// The package simulation defines the Simulation interface for running simulations.
package simulation

import (
	"context"
	"log/slog"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/airport"
)

// PreSimulationPlugin defines a plugin that modifies the airport configuration before the simulation runs.
type PreSimulationPlugin interface {
	Apply(airport.Airport) airport.Airport
}

// Simulation represents a simulation that can be run.
type Simulation struct {
	airport              airport.Airport       // The airport to simulate.
	logger               slog.Logger           // The logger to use for logging.
	preSimulationPlugins []PreSimulationPlugin // Pre-simulation plugins to modify the airport configuration.
}

// NewSimulation creates a new Simulation instance.
func NewSimulation(airport airport.Airport, logger slog.Logger) Simulation {
	return Simulation{
		airport:              airport,
		logger:               logger,
		preSimulationPlugins: []PreSimulationPlugin{},
	}
}

// AddPreSimulationPlugin adds a pre-simulation plugin to the simulation.
func (s Simulation) AddPreSimulationPlugin(plugin PreSimulationPlugin) Simulation {
	s.preSimulationPlugins = append(s.preSimulationPlugins, plugin)
	return s
}

func (s Simulation) Run(ctx context.Context) (float32, error) {
	// Apply pre-simulation plugins
	for _, plugin := range s.preSimulationPlugins {
		s.airport = plugin.Apply(s.airport)
	}

	// Here you would implement the actual simulation logic.
	s.logger.InfoContext(ctx, "Running simulation for airport", "airport", s.airport.Name)
	basicSim := NewBasicSim(s.airport, &s.logger)
	return basicSim.Run(ctx)
}
