package policy

import (
	"context"
	"fmt"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/event"
)

// RotationStrategy defines how runways are rotated to minimize noise impact.
type RotationStrategy int

const (
	// NoRotation means runways are used based on wind/efficiency only
	NoRotation RotationStrategy = iota

	// TimeBasedRotation rotates runway usage at fixed time intervals
	TimeBasedRotation

	// BalancedRotation attempts to balance usage across all runways
	BalancedRotation

	// NoiseOptimizedRotation rotates to minimize noise impact on communities
	NoiseOptimizedRotation
)

// String returns the string representation of the rotation strategy.
func (rs RotationStrategy) String() string {
	switch rs {
	case NoRotation:
		return "NoRotation"
	case TimeBasedRotation:
		return "TimeBasedRotation"
	case BalancedRotation:
		return "BalancedRotation"
	case NoiseOptimizedRotation:
		return "NoiseOptimizedRotation"
	default:
		return "Unknown"
	}
}

// RotationPolicyConfiguration holds configuration for runway rotation policies.
type RotationPolicyConfiguration struct {
	efficiencyMap map[RotationStrategy]float32
}

// NewDefaultRotationPolicyConfiguration creates a new default rotation policy configuration
func NewDefaultRotationPolicyConfiguration() *RotationPolicyConfiguration {
	return &RotationPolicyConfiguration{
		efficiencyMap: map[RotationStrategy]float32{
			NoRotation:             1.0,
			TimeBasedRotation:      0.95,
			BalancedRotation:       0.90,
			NoiseOptimizedRotation: 0.80,
		},
	}
}

// NewRotationPolicyConfiguration creates a new rotation policy configuration
func NewRotationPolicyConfiguration(efficiencyMap map[RotationStrategy]float32) *RotationPolicyConfiguration {
	return &RotationPolicyConfiguration{
		efficiencyMap: efficiencyMap,
	}
}

// RunwayRotationPolicy implements runway rotation strategies to distribute
// aircraft movements across different runways over time.
type RunwayRotationPolicy struct {
	strategy RotationStrategy             // The selected rotation strategy
	config   *RotationPolicyConfiguration // Configuration for efficiency adjustments
}

// NewRunwayRotationPolicy creates a new runway rotation policy.
func NewRunwayRotationPolicy(strategy RotationStrategy, config *RotationPolicyConfiguration) *RunwayRotationPolicy {
	return &RunwayRotationPolicy{
		strategy: strategy,
		config:   config,
	}
}

// NewDefaultRunwayRotationPolicy creates a new runway rotation policy with the default configuration
func NewDefaultRunwayRotationPolicy(strategy RotationStrategy) *RunwayRotationPolicy {
	return &RunwayRotationPolicy{
		strategy: strategy,
		config:   NewDefaultRotationPolicyConfiguration(),
	}
}

// Name returns the policy name.
func (p *RunwayRotationPolicy) Name() string {
	return fmt.Sprintf("RunwayRotationPolicy(%s)", p.strategy.String())
}

// GenerateEvents generates a rotation change event at the start of the simulation.
// Different strategies affect capacity by applying efficiency multipliers.
// Rotation strategies introduce overhead and constraints that reduce theoretical maximum capacity.
func (p *RunwayRotationPolicy) GenerateEvents(ctx context.Context, world EventWorld) error {
	startTime := world.GetStartTime()

	// Get efficiency multiplier based on rotation strategy
	var efficiencyMultiplier float32
	switch p.strategy {
	case NoRotation:
		// No modification needed - use runways as efficiently as possible
		// Efficiency: 100% (no penalty)
		efficiencyMultiplier = p.config.efficiencyMap[NoRotation]

	case TimeBasedRotation:
		// Time-based rotation introduces small overhead during transition periods
		// when runways are switched. This accounts for the time to communicate
		// the change to air traffic control and pilots.
		// Efficiency: 95% (5% capacity reduction)
		efficiencyMultiplier = p.config.efficiencyMap[TimeBasedRotation]

	case BalancedRotation:
		// Balanced rotation distributes usage across all runways equally,
		// which may not always align with optimal wind conditions or traffic flow.
		// This introduces moderate efficiency penalties.
		// Efficiency: 90% (10% capacity reduction)
		efficiencyMultiplier = p.config.efficiencyMap[BalancedRotation]

	case NoiseOptimizedRotation:
		// Noise-optimized rotation prioritizes minimizing community noise impact
		// over operational efficiency. This typically requires using less efficient
		// runway configurations and more restrictive flight paths.
		// Efficiency: 80% (20% capacity reduction)
		efficiencyMultiplier = p.config.efficiencyMap[NoiseOptimizedRotation]

	default:
		return fmt.Errorf("unknown rotation strategy: %v", p.strategy)
	}

	// Schedule a rotation change event at the start of the simulation
	// This sets the efficiency multiplier for the entire simulation period
	world.ScheduleEvent(event.NewRotationChangeEvent(efficiencyMultiplier, startTime))

	return nil
}
