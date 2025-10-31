package policy

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// Common errors for curfew policy validation
var (
	// ErrInvalidCurfewTime indicates the curfew time range is invalid
	ErrInvalidCurfewTime = errors.New("curfew end time must be after start time")

	// ErrCurfewTooLong indicates the curfew duration exceeds reasonable limits
	ErrCurfewTooLong = errors.New("curfew duration exceeds maximum allowed duration")
)

const (
	// MaxCurfewDuration defines the maximum allowed curfew duration (30 days)
	// This prevents misconfiguration where extremely long curfews would make the simulation invalid
	MaxCurfewDuration = 30 * 24 * time.Hour
)

// CurfewPolicy restricts airport operations during specified time ranges.
// It reduces the effective operating hours of the airport.
type CurfewPolicy struct {
	startTime time.Time // Start of curfew period
	endTime   time.Time // End of curfew period
}

// NewCurfewPolicy creates a new curfew policy with validation.
// Returns an error if the time range is invalid.
func NewCurfewPolicy(startTime, endTime time.Time) (*CurfewPolicy, error) {
	// Validate that end time is after start time
	if !endTime.After(startTime) {
		return nil, ErrInvalidCurfewTime
	}

	// Validate that the duration is reasonable
	duration := endTime.Sub(startTime)
	if duration > MaxCurfewDuration {
		return nil, ErrCurfewTooLong
	}

	return &CurfewPolicy{
		startTime: startTime,
		endTime:   endTime,
	}, nil
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
		// If operating hours not yet set, assume 24/7 operation
		currentHours = HoursPerYear
	}

	// Calculate daily curfew as a percentage of 24 hours
	dailyCurfewHours := curfewDuration
	if dailyCurfewHours > HoursPerDay {
		dailyCurfewHours = curfewDuration - float64(int(curfewDuration/HoursPerDay)*HoursPerDay)
	}

	// Reduce annual operating hours by daily curfew * days per year
	reducedHours := max(currentHours-float32(dailyCurfewHours*DaysPerYear), 0)

	simState.SetOperatingHours(reducedHours)

	logger.InfoContext(ctx, "Curfew policy applied",
		"curfew_duration_hours", curfewDuration,
		"daily_curfew_hours", dailyCurfewHours,
		"hours_before", currentHours,
		"hours_after", reducedHours)

	return nil
}
