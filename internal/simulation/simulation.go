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

// Policy defines a runtime policy that affects simulation behavior during execution.
type Policy interface {
	Name() string
	Apply(ctx context.Context, state any, logger *slog.Logger) error
}

// Type aliases for convenience - expose policy package types
type (
	MaintenanceSchedule = policy.MaintenanceSchedule
	RotationStrategy    = policy.RotationStrategy
)

// Rotation strategy constants
const (
	NoRotation             = policy.NoRotation
	TimeBasedRotation      = policy.TimeBasedRotation
	BalancedRotation       = policy.BalancedRotation
	NoiseOptimizedRotation = policy.NoiseOptimizedRotation
)

// SimulationState represents the mutable state during simulation execution.
type SimulationState struct {
	Airport          airport.Airport  // The airport being simulated
	CurrentTime      time.Time        // Current simulation time
	AvailableRunways []airport.Runway // Runways currently available
	TotalMovements   float32          // Total movements processed
	OperatingHours   float32          // Total hours of operation
}

// Simulation represents a simulation that can be run.
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

func (s *Simulation) Run(ctx context.Context) (float32, error) {
	// Apply pre-simulation plugins
	for _, plugin := range s.preSimulationPlugins {
		s.airport = plugin.Apply(s.airport)
	}

	// Initialize simulation state
	state := &SimulationState{
		Airport:          s.airport,
		CurrentTime:      time.Now(),
		AvailableRunways: s.airport.Runways,
		TotalMovements:   0,
		OperatingHours:   0,
	}

	// Apply runtime policies
	for _, policy := range s.policies {
		s.logger.InfoContext(ctx, "Applying policy", "policy", policy.Name())
		if err := policy.Apply(ctx, state, s.logger); err != nil {
			s.logger.ErrorContext(ctx, "Failed to apply policy", "policy", policy.Name(), "error", err)
			return 0, err
		}
	}

	// Run the simulation engine with the policy-modified state
	s.logger.InfoContext(ctx, "Running simulation for airport", "airport", s.airport.Name)
	engine := NewEngine(s.logger)
	return engine.Calculate(ctx, state)
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

// RunwayRotationPolicy adds a runway rotation policy that implements rotation strategies.
func (s *Simulation) RunwayRotationPolicy(strategy RotationStrategy) *Simulation {
	p := policy.NewDefaultRunwayRotationPolicy(strategy)
	return s.AddPolicy(p)
}

// Helper methods for SimulationState to support policy operations
func (ss *SimulationState) GetOperatingHours() float32 {
	return ss.OperatingHours
}

func (ss *SimulationState) SetOperatingHours(hours float32) {
	ss.OperatingHours = hours
}

func (ss *SimulationState) GetAvailableRunways() []airport.Runway {
	return ss.AvailableRunways
}

func (ss *SimulationState) SetAvailableRunways(runways []airport.Runway) {
	ss.AvailableRunways = runways
}
