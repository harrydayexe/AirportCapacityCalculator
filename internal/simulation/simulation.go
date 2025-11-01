// The package simulation defines the Simulation interface for running simulations.
package simulation

import (
	"context"
	"log/slog"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/airport"
	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/policy"
)

// PreSimulationPlugin defines a plugin that modifies the airport configuration before the simulation runs.
type PreSimulationPlugin interface {
	Apply(airport.Airport) airport.Airport
}

// Policy defines a runtime policy that generates events for the event-driven simulation.
type Policy interface {
	Name() string
	GenerateEvents(ctx context.Context, world policy.EventWorld) error
}

// Type aliases for convenience - expose policy package types
type (
	MaintenanceSchedule           = policy.MaintenanceSchedule
	IntelligentMaintenanceSchedule = policy.IntelligentMaintenanceSchedule
	PeakHours                     = policy.PeakHours
	RotationStrategy              = policy.RotationStrategy
)

// Rotation strategy constants
const (
	NoRotation             = policy.NoRotation
	TimeBasedRotation      = policy.TimeBasedRotation
	BalancedRotation       = policy.BalancedRotation
	NoiseOptimizedRotation = policy.NoiseOptimizedRotation
)

// Simulation represents an event-driven simulation that can be run.
type Simulation struct {
	airport              airport.Airport       // The airport to simulate.
	logger               *slog.Logger          // The logger to use for logging.
	preSimulationPlugins []PreSimulationPlugin // Pre-simulation plugins to modify the airport configuration.
	policies             []Policy              // Runtime policies affecting simulation behavior.
}

// NewSimulation creates a new Simulation instance.
func NewSimulation(airport airport.Airport, logger *slog.Logger) *Simulation {
	return &Simulation{
		airport:              airport,
		logger:               logger,
		preSimulationPlugins: []PreSimulationPlugin{},
		policies:             []Policy{},
	}
}

// AddPreSimulationPlugin adds a pre-simulation plugin to the simulation.
func (s *Simulation) AddPreSimulationPlugin(plugin PreSimulationPlugin) *Simulation {
	s.preSimulationPlugins = append(s.preSimulationPlugins, plugin)
	return s
}

// Run executes the event-driven simulation.
func (s *Simulation) Run(ctx context.Context) (float32, error) {
	// Apply pre-simulation plugins
	for _, plugin := range s.preSimulationPlugins {
		s.airport = plugin.Apply(s.airport)
	}

	// Create simulation world
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := startTime.AddDate(1, 0, 0) // One year simulation

	world := NewWorld(s.airport, startTime, endTime)

	s.logger.InfoContext(ctx, "Starting event-driven simulation",
		"airport", s.airport.Name,
		"startTime", startTime,
		"endTime", endTime)

	// Let policies generate events
	for _, policy := range s.policies {
		s.logger.InfoContext(ctx, "Generating events for policy", "policy", policy.Name())
		if err := policy.GenerateEvents(ctx, world); err != nil {
			s.logger.ErrorContext(ctx, "Failed to generate events",
				"policy", policy.Name(),
				"error", err)
			return 0, err
		}
	}

	s.logger.InfoContext(ctx, "Events generated",
		"totalEvents", world.Events.Len())

	// Run event-driven simulation
	engine := NewEngine(s.logger)
	return engine.Calculate(ctx, world)
}

// AddPolicy adds a runtime policy to the simulation.
func (s *Simulation) AddPolicy(policy Policy) *Simulation {
	s.policies = append(s.policies, policy)
	return s
}

// AddCurfewPolicy adds a curfew policy that restricts airport operations during specified hours.
// Returns an error if the curfew time range is invalid.
func (s *Simulation) AddCurfewPolicy(startTime, endTime time.Time) (*Simulation, error) {
	p, err := policy.NewCurfewPolicy(startTime, endTime)
	if err != nil {
		return nil, err
	}
	return s.AddPolicy(p), nil
}

// AddMaintenancePolicy adds a maintenance policy that schedules runway maintenance.
func (s *Simulation) AddMaintenancePolicy(schedule MaintenanceSchedule) *Simulation {
	p := policy.NewMaintenancePolicy(schedule)
	return s.AddPolicy(p)
}

// AddIntelligentMaintenancePolicy adds an intelligent maintenance policy that optimizes
// maintenance scheduling by coordinating with curfews, avoiding peak hours, and ensuring
// minimum operational runway capacity.
func (s *Simulation) AddIntelligentMaintenancePolicy(schedule IntelligentMaintenanceSchedule) (*Simulation, error) {
	p, err := policy.NewIntelligentMaintenancePolicy(schedule)
	if err != nil {
		return nil, err
	}
	return s.AddPolicy(p), nil
}

// RunwayRotationPolicy adds a runway rotation policy that implements rotation strategies.
func (s *Simulation) RunwayRotationPolicy(strategy RotationStrategy) *Simulation {
	p := policy.NewDefaultRunwayRotationPolicy(strategy)
	return s.AddPolicy(p)
}
