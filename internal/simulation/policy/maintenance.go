package policy

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/airport"
)

// MaintenanceSchedule defines a maintenance schedule for runways.
type MaintenanceSchedule struct {
	RunwayDesignations []string      // Runway identifiers to maintain
	Duration           time.Duration // Duration of maintenance window
	Frequency          time.Duration // How often maintenance occurs
}

// MaintenancePolicy schedules runway maintenance that temporarily removes runways from operation.
type MaintenancePolicy struct {
	schedule MaintenanceSchedule
}

// NewMaintenancePolicy creates a new maintenance policy.
func NewMaintenancePolicy(schedule MaintenanceSchedule) *MaintenancePolicy {
	return &MaintenancePolicy{
		schedule: schedule,
	}
}

// Name returns the policy name.
func (p *MaintenancePolicy) Name() string {
	return "MaintenancePolicy"
}

// Apply applies the maintenance policy to the simulation state.
// It removes runways from the available runway list during maintenance windows.
func (p *MaintenancePolicy) Apply(ctx context.Context, state interface{}, logger *slog.Logger) error {
	// Type assertion to SimulationState interface
	simState, ok := state.(interface {
		GetAvailableRunways() []airport.Runway
		SetAvailableRunways([]airport.Runway)
		GetOperatingHours() float32
		SetOperatingHours(float32)
	})

	if !ok {
		return fmt.Errorf("invalid state type for MaintenancePolicy")
	}

	logger.DebugContext(ctx, "Applying maintenance policy",
		"runways", p.schedule.RunwayDesignations,
		"duration", p.schedule.Duration,
		"frequency", p.schedule.Frequency)

	// Calculate maintenance downtime
	// Number of maintenance windows per year
	maintenanceWindows := float64(YearDuration) / float64(p.schedule.Frequency)

	// Total hours of maintenance per year
	maintenanceHoursPerYear := maintenanceWindows * p.schedule.Duration.Hours()

	// For now, we'll reduce the overall operating efficiency
	// A more sophisticated implementation would track which specific runways
	// are unavailable at specific times
	totalHours := simState.GetOperatingHours()
	if totalHours == 0 {
		totalHours = HoursPerYear // Default to full year
	}

	// This is a simplified model - in reality, you'd want to track
	// per-runway availability over time

	// Reduce operating hours by maintenance downtime
	reducedHours := max(totalHours-float32(maintenanceHoursPerYear), 0)
	simState.SetOperatingHours(reducedHours)

	logger.InfoContext(ctx, "Maintenance policy applied",
		"runways", p.schedule.RunwayDesignations,
		"maintenance_windows_per_year", maintenanceWindows,
		"maintenance_hours_per_year", maintenanceHoursPerYear,
		"hours_before", totalHours,
		"hours_after", reducedHours)

	return nil
}
