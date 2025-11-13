package event

import (
	"context"
	"time"
)

// TaxiTimeAdjustmentEvent represents taxi time overhead being applied to capacity calculations.
type TaxiTimeAdjustmentEvent struct {
	totalTaxiTimeOverhead time.Duration
	timestamp             time.Time
}

// NewTaxiTimeAdjustmentEvent creates a new taxi time adjustment event.
func NewTaxiTimeAdjustmentEvent(totalTaxiTimeOverhead time.Duration, timestamp time.Time) *TaxiTimeAdjustmentEvent {
	return &TaxiTimeAdjustmentEvent{
		totalTaxiTimeOverhead: totalTaxiTimeOverhead,
		timestamp:             timestamp,
	}
}

// Time returns when the adjustment is applied.
func (e *TaxiTimeAdjustmentEvent) Time() time.Time {
	return e.timestamp
}

// Type returns the event type.
func (e *TaxiTimeAdjustmentEvent) Type() EventType {
	return TaxiTimeAdjustmentType
}

// TotalTaxiTimeOverhead returns the total taxi time overhead per aircraft cycle.
func (e *TaxiTimeAdjustmentEvent) TotalTaxiTimeOverhead() time.Duration {
	return e.totalTaxiTimeOverhead
}

// Apply sets the taxi time overhead in the world state.
func (e *TaxiTimeAdjustmentEvent) Apply(ctx context.Context, world WorldState) error {
	return world.SetTaxiTimeOverhead(e.totalTaxiTimeOverhead)
}
