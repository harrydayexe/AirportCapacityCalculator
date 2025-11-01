package policy

import (
	"context"
	"fmt"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/event"
)

// GateCapacityConstraint defines gate capacity limitations.
type GateCapacityConstraint struct {
	TotalGates          int           // Total number of gates at the airport
	AverageTurnaroundTime time.Duration // Average time aircraft occupies a gate
}

// GateCapacityPolicy models the constraint that gate availability places on sustained throughput.
// When gates are fully utilized, they limit the airport's ability to accept new arrivals,
// effectively capping the sustained capacity below what runways could theoretically handle.
type GateCapacityPolicy struct {
	constraint GateCapacityConstraint
}

// NewGateCapacityPolicy creates a new gate capacity policy.
func NewGateCapacityPolicy(constraint GateCapacityConstraint) (*GateCapacityPolicy, error) {
	if constraint.TotalGates <= 0 {
		return nil, fmt.Errorf("total gates must be positive, got %d", constraint.TotalGates)
	}
	if constraint.AverageTurnaroundTime <= 0 {
		return nil, fmt.Errorf("average turnaround time must be positive, got %v", constraint.AverageTurnaroundTime)
	}

	return &GateCapacityPolicy{
		constraint: constraint,
	}, nil
}

// Name returns the policy name.
func (p *GateCapacityPolicy) Name() string {
	return "GateCapacityPolicy"
}

// GenerateEvents generates a gate capacity constraint event at simulation start.
// This event applies a capacity multiplier that represents the limitation gates
// place on sustained throughput.
//
// The multiplier is calculated as:
// - Sustained arrival rate = gates / turnaround_time
// - This becomes a cap on total movements if it's lower than runway capacity
//
// Note: This is a simplified model for v0.3.0. Future versions may implement
// more sophisticated gate utilization tracking with per-flight occupancy.
func (p *GateCapacityPolicy) GenerateEvents(ctx context.Context, world EventWorld) error {
	startTime := world.GetStartTime()

	// Calculate the gate-limited sustained capacity
	// If we have N gates and average turnaround of T hours,
	// we can handle at most N/T arrivals per hour sustained
	turnaroundHours := p.constraint.AverageTurnaroundTime.Hours()
	sustainedArrivalsPerHour := float32(p.constraint.TotalGates) / float32(turnaroundHours)

	// Since movements include both arrivals and departures, and in steady state
	// they're equal, the total movement capacity is 2x arrivals
	gateConstrainedMovementsPerHour := sustainedArrivalsPerHour * 2

	// Convert to movements per second for consistency with runway separation
	gateConstrainedMovementsPerSecond := gateConstrainedMovementsPerHour / 3600.0

	// Schedule the gate capacity constraint event
	world.ScheduleEvent(event.NewGateCapacityConstraintEvent(
		gateConstrainedMovementsPerSecond,
		startTime,
	))

	return nil
}
