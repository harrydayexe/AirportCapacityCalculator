package policy

import (
	"context"
	"fmt"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/event"
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

// GenerateEvents generates maintenance start and end events for each runway according to the schedule.
// Maintenance windows are distributed evenly across the simulation period.
func (p *MaintenancePolicy) GenerateEvents(ctx context.Context, world EventWorld) error {
	startTime := world.GetStartTime()
	endTime := world.GetEndTime()
	simulationDuration := endTime.Sub(startTime)

	// Calculate number of maintenance windows for the simulation period
	maintenanceWindows := int(simulationDuration / p.schedule.Frequency)
	if maintenanceWindows == 0 {
		maintenanceWindows = 1 // At least one maintenance window
	}

	// Get all runway IDs from world
	allRunwayIDs := world.GetRunwayIDs()

	// Generate maintenance events for each specified runway
	for _, runwayDesignation := range p.schedule.RunwayDesignations {
		// Verify runway exists
		runwayExists := false
		for _, id := range allRunwayIDs {
			if id == runwayDesignation {
				runwayExists = true
				break
			}
		}

		if !runwayExists {
			return fmt.Errorf("runway %s not found in airport", runwayDesignation)
		}

		// Schedule maintenance windows evenly across the year
		currentTime := startTime
		for i := 0; i < maintenanceWindows; i++ {
			// Schedule maintenance start event
			maintenanceStart := currentTime
			if maintenanceStart.Before(endTime) {
				world.ScheduleEvent(event.NewRunwayMaintenanceStartEvent(runwayDesignation, maintenanceStart))
			}

			// Schedule maintenance end event
			maintenanceEnd := maintenanceStart.Add(p.schedule.Duration)
			if maintenanceEnd.Before(endTime) {
				world.ScheduleEvent(event.NewRunwayMaintenanceEndEvent(runwayDesignation, maintenanceEnd))
			}

			// Move to next maintenance window
			currentTime = currentTime.Add(p.schedule.Frequency)
		}
	}

	return nil
}
