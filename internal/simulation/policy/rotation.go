package policy

import (
	"context"
	"fmt"
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

// RunwayRotationPolicy implements runway rotation strategies to distribute
// aircraft movements across different runways over time.
type RunwayRotationPolicy struct {
	strategy RotationStrategy
}

// NewRunwayRotationPolicy creates a new runway rotation policy.
func NewRunwayRotationPolicy(strategy RotationStrategy) *RunwayRotationPolicy {
	return &RunwayRotationPolicy{
		strategy: strategy,
	}
}

// Name returns the policy name.
func (p *RunwayRotationPolicy) Name() string {
	return fmt.Sprintf("RunwayRotationPolicy(%s)", p.strategy.String())
}

// Apply applies the runway rotation policy to the simulation state.
// Different strategies affect capacity by applying efficiency multipliers to operating hours.
// Rotation strategies introduce overhead and constraints that reduce theoretical maximum capacity.
func (p *RunwayRotationPolicy) Apply(ctx context.Context, state interface{}) error {
	// Type assertion to get state with operating hours
	simState, ok := state.(interface {
		GetOperatingHours() float32
		SetOperatingHours(float32)
	})

	if !ok {
		return fmt.Errorf("invalid state type for RunwayRotationPolicy")
	}

	// Get current operating hours (if not set, assume 24/7 operation)
	currentHours := simState.GetOperatingHours()
	if currentHours == 0 {
		currentHours = 8760 // 365 days * 24 hours
	}

	// Apply efficiency multiplier based on rotation strategy
	var efficiencyMultiplier float32
	switch p.strategy {
	case NoRotation:
		// No modification needed - use runways as efficiently as possible
		// Efficiency: 100% (no penalty)
		efficiencyMultiplier = 1.0

	case TimeBasedRotation:
		// Time-based rotation introduces small overhead during transition periods
		// when runways are switched. This accounts for the time to communicate
		// the change to air traffic control and pilots.
		// Efficiency: 95% (5% capacity reduction)
		efficiencyMultiplier = 0.95

	case BalancedRotation:
		// Balanced rotation distributes usage across all runways equally,
		// which may not always align with optimal wind conditions or traffic flow.
		// This introduces moderate efficiency penalties.
		// Efficiency: 90% (10% capacity reduction)
		efficiencyMultiplier = 0.90

	case NoiseOptimizedRotation:
		// Noise-optimized rotation prioritizes minimizing community noise impact
		// over operational efficiency. This typically requires using less efficient
		// runway configurations and more restrictive flight paths.
		// Efficiency: 80% (20% capacity reduction)
		efficiencyMultiplier = 0.80

	default:
		return fmt.Errorf("unknown rotation strategy: %v", p.strategy)
	}

	// Apply the efficiency multiplier to operating hours
	// This effectively reduces capacity by the specified percentage
	adjustedHours := currentHours * efficiencyMultiplier
	simState.SetOperatingHours(adjustedHours)

	return nil
}
