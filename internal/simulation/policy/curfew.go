package policy

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// CurfewPolicy restricts airport operations during specified time ranges.
// It reduces the effective operating hours of the airport.
type CurfewPolicy struct {
	startTime time.Time // Start of curfew period
	endTime   time.Time // End of curfew period
}

// NewCurfewPolicy creates a new curfew policy.
func NewCurfewPolicy(startTime, endTime time.Time) *CurfewPolicy {
	return &CurfewPolicy{
		startTime: startTime,
		endTime:   endTime,
	}
}

// Name returns the policy name.
func (p *CurfewPolicy) Name() string {
	return "CurfewPolicy"
}

// Apply applies the curfew policy to the simulation state.
// It calculates the reduced operating hours based on the curfew period.
func (p *CurfewPolicy) Apply(ctx context.Context, state any, logger *slog.Logger) error {
	// Type assertion to SimulationState interface
	simState, ok := state.(interface {
		GetOperatingHours() float32
		SetOperatingHours(float32)
	})

	if !ok {
		return fmt.Errorf("invalid state type for CurfewPolicy")
	}

	logger.DebugContext(ctx, "Applying curfew policy",
		"start_time", p.startTime,
		"end_time", p.endTime)

	// Calculate curfew duration in hours
	curfewDuration := p.endTime.Sub(p.startTime).Hours()

	// Reduce operating hours by curfew duration
	currentHours := simState.GetOperatingHours()
	if currentHours == 0 {
		// If operating hours not yet set, assume 24/7 operation (8760 hours per year)
		currentHours = 8760
	}

	// Calculate daily curfew as a percentage of 24 hours
	dailyCurfewHours := curfewDuration
	if dailyCurfewHours > 24 {
		dailyCurfewHours = curfewDuration - float64(int(curfewDuration/24)*24)
	}

	// Reduce annual operating hours by daily curfew * 365 days
	reducedHours := max(currentHours-float32(dailyCurfewHours*365), 0)

	simState.SetOperatingHours(reducedHours)

	logger.InfoContext(ctx, "Curfew policy applied",
		"curfew_duration_hours", curfewDuration,
		"daily_curfew_hours", dailyCurfewHours,
		"hours_before", currentHours,
		"hours_after", reducedHours)

	return nil
}
