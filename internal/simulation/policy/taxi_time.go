package policy

import (
	"context"
	"fmt"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/event"
)

// TaxiTimeConfiguration defines taxi time parameters.
type TaxiTimeConfiguration struct {
	AverageTaxiInTime  time.Duration // Average time from runway to gate
	AverageTaxiOutTime time.Duration // Average time from gate to runway
}

// TaxiTimePolicy models the impact of taxi time on airport capacity.
// Taxi time affects:
// - Effective gate occupancy (aircraft occupy gates longer due to taxi time)
// - Runway exit efficiency (faster taxi clears runway area sooner)
//
// For v0.3.0, this policy primarily adjusts gate capacity by accounting for
// taxi time overhead in the effective turnaround time.
type TaxiTimePolicy struct {
	config TaxiTimeConfiguration
}

// NewTaxiTimePolicy creates a new taxi time policy.
func NewTaxiTimePolicy(config TaxiTimeConfiguration) (*TaxiTimePolicy, error) {
	if config.AverageTaxiInTime < 0 {
		return nil, fmt.Errorf("average taxi-in time cannot be negative: %v", config.AverageTaxiInTime)
	}
	if config.AverageTaxiOutTime < 0 {
		return nil, fmt.Errorf("average taxi-out time cannot be negative: %v", config.AverageTaxiOutTime)
	}

	return &TaxiTimePolicy{
		config: config,
	}, nil
}

// Name returns the policy name.
func (p *TaxiTimePolicy) Name() string {
	return "TaxiTimePolicy"
}

// GenerateEvents generates a taxi time adjustment event at simulation start.
//
// Taxi time extends the effective time an aircraft occupies the airport system:
// - Arrival: lands, taxis in (taxi-in time), occupies gate, taxis out (taxi-out time), departs
// - Total taxi overhead = taxi-in + taxi-out
//
// This overhead reduces sustainable throughput by extending the effective
// turnaround time. The policy generates an event that can be used by the
// engine to adjust capacity calculations.
//
// Note: This is a simplified model. Future versions may implement:
// - Taxiway capacity constraints (max aircraft on taxiways)
// - Runway exit efficiency modeling
// - Hot spot and conflict point detection
func (p *TaxiTimePolicy) GenerateEvents(ctx context.Context, world EventWorld) error {
	startTime := world.GetStartTime()

	// Total taxi time overhead per aircraft cycle
	totalTaxiTimeOverhead := p.config.AverageTaxiInTime + p.config.AverageTaxiOutTime

	// Generate taxi time adjustment event
	world.ScheduleEvent(event.NewTaxiTimeAdjustmentEvent(
		totalTaxiTimeOverhead,
		startTime,
	))

	return nil
}
