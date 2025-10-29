package policy

import (
	"context"
	"fmt"
	"time"
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
func (p *MaintenancePolicy) Apply(ctx context.Context, state interface{}) error {
	// Type assertion to SimulationState interface
	simState, ok := state.(interface {
		GetAvailableRunways() interface{}
		SetAvailableRunways(interface{})
		GetOperatingHours() float32
	})

	if !ok {
		return fmt.Errorf("invalid state type for MaintenancePolicy")
	}

	// Calculate maintenance downtime
	// Number of maintenance windows per year
	yearDuration := 365 * 24 * time.Hour
	maintenanceWindows := float64(yearDuration) / float64(p.schedule.Frequency)

	// Total hours of maintenance per year
	maintenanceHoursPerYear := maintenanceWindows * p.schedule.Duration.Hours()

	// For now, we'll reduce the overall operating efficiency
	// A more sophisticated implementation would track which specific runways
	// are unavailable at specific times
	totalHours := simState.GetOperatingHours()
	if totalHours == 0 {
		totalHours = 8760 // Default to full year
	}

	// This is a simplified model - in reality, you'd want to track
	// per-runway availability over time
	_ = maintenanceHoursPerYear

	return nil
}
