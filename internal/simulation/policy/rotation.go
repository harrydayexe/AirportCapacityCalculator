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
// Different strategies will affect how runway capacity is calculated.
func (p *RunwayRotationPolicy) Apply(ctx context.Context, state interface{}) error {
	// Type assertion to get state with runway information
	_, ok := state.(interface {
		GetAvailableRunways() interface{}
	})

	if !ok {
		return fmt.Errorf("invalid state type for RunwayRotationPolicy")
	}

	// The actual impact of rotation strategies would be implemented here
	// For now, this is a placeholder that sets up the strategy
	switch p.strategy {
	case NoRotation:
		// No modification needed - use runways as efficiently as possible
		return nil

	case TimeBasedRotation:
		// Would implement time-based rotation logic
		// This might reduce effective capacity slightly due to transition overhead
		return nil

	case BalancedRotation:
		// Would implement balanced usage across runways
		// Might have small efficiency penalty
		return nil

	case NoiseOptimizedRotation:
		// Would implement noise-optimized rotation
		// Typically has the highest efficiency penalty
		return nil

	default:
		return fmt.Errorf("unknown rotation strategy: %v", p.strategy)
	}
}
