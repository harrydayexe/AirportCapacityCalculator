package policy

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/event"
)

// TimeWindow represents a time period.
type TimeWindow struct {
	Start time.Time
	End   time.Time
}

// IntelligentMaintenanceSchedule defines an intelligent maintenance schedule that coordinates with operational constraints.
type IntelligentMaintenanceSchedule struct {
	RunwayDesignations       []string      // Runway identifiers to maintain
	Duration                 time.Duration // Duration of maintenance window
	Frequency                time.Duration // How often maintenance must occur
	MinimumOperationalRunways int           // Minimum runways that must remain operational (default: 1)
	CurfewStart              *time.Time    // Optional: daily curfew start time (for coordination)
	CurfewEnd                *time.Time    // Optional: daily curfew end time
}

// IntelligentMaintenancePolicy schedules runway maintenance intelligently by:
// - Preferring maintenance during or adjacent to curfew periods
// - Coordinating across runways to maintain minimum operational capacity
type IntelligentMaintenancePolicy struct {
	schedule IntelligentMaintenanceSchedule
}

// NewIntelligentMaintenancePolicy creates a new intelligent maintenance policy.
func NewIntelligentMaintenancePolicy(schedule IntelligentMaintenanceSchedule) (*IntelligentMaintenancePolicy, error) {
	// Set defaults
	if schedule.MinimumOperationalRunways <= 0 {
		schedule.MinimumOperationalRunways = 1
	}

	return &IntelligentMaintenancePolicy{
		schedule: schedule,
	}, nil
}

// Name returns the policy name.
func (p *IntelligentMaintenancePolicy) Name() string {
	return "IntelligentMaintenancePolicy"
}

// maintenanceWindow represents a scheduled maintenance period for a runway.
type maintenanceWindow struct {
	RunwayID string
	Start    time.Time
	End      time.Time
}

// GenerateEvents generates intelligently scheduled maintenance events.
func (p *IntelligentMaintenancePolicy) GenerateEvents(ctx context.Context, world EventWorld) error {
	startTime := world.GetStartTime()
	endTime := world.GetEndTime()
	simulationDuration := endTime.Sub(startTime)

	// Calculate number of maintenance windows needed
	maintenanceWindows := int(simulationDuration / p.schedule.Frequency)
	if maintenanceWindows == 0 {
		maintenanceWindows = 1
	}

	// Verify all runways exist
	allRunwayIDs := world.GetRunwayIDs()
	for _, runwayDesignation := range p.schedule.RunwayDesignations {
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
	}

	// Build curfew windows for the entire simulation period
	curfewWindows := p.buildCurfewWindows(startTime, endTime)

	// Track maintenance schedules for runway coordination
	scheduledMaintenance := []maintenanceWindow{}

	// Schedule maintenance for each runway
	for runwayIdx, runwayDesignation := range p.schedule.RunwayDesignations {
		// Stagger start times to distribute maintenance across runways
		offset := time.Duration(runwayIdx) * (p.schedule.Frequency / time.Duration(len(p.schedule.RunwayDesignations)))
		currentTime := startTime.Add(offset)

		for i := 0; i < maintenanceWindows; i++ {
			// Find optimal maintenance window
			maintenanceStart := p.findOptimalWindow(
				currentTime,
				endTime,
				curfewWindows,
				scheduledMaintenance,
			)

			// If we couldn't find an optimal window, use current time
			if maintenanceStart.IsZero() {
				maintenanceStart = currentTime
			}

			// Ensure we don't exceed simulation end
			if maintenanceStart.Add(p.schedule.Duration).After(endTime) {
				break
			}

			// Schedule maintenance events
			maintenanceEnd := maintenanceStart.Add(p.schedule.Duration)
			world.ScheduleEvent(event.NewRunwayMaintenanceStartEvent(runwayDesignation, maintenanceStart))
			world.ScheduleEvent(event.NewRunwayMaintenanceEndEvent(runwayDesignation, maintenanceEnd))

			// Track this maintenance window
			scheduledMaintenance = append(scheduledMaintenance, maintenanceWindow{
				RunwayID: runwayDesignation,
				Start:    maintenanceStart,
				End:      maintenanceEnd,
			})

			// Move to next maintenance cycle
			currentTime = currentTime.Add(p.schedule.Frequency)
		}
	}

	return nil
}

// buildCurfewWindows builds all curfew time windows for the simulation period.
func (p *IntelligentMaintenancePolicy) buildCurfewWindows(startTime, endTime time.Time) []TimeWindow {
	if p.schedule.CurfewStart == nil || p.schedule.CurfewEnd == nil {
		return nil
	}

	windows := []TimeWindow{}
	currentDate := startTime

	curfewStartHour, curfewStartMinute := p.schedule.CurfewStart.Hour(), p.schedule.CurfewStart.Minute()
	curfewEndHour, curfewEndMinute := p.schedule.CurfewEnd.Hour(), p.schedule.CurfewEnd.Minute()

	for currentDate.Before(endTime) {
		curfewStart := time.Date(
			currentDate.Year(), currentDate.Month(), currentDate.Day(),
			curfewStartHour, curfewStartMinute, 0, 0,
			currentDate.Location(),
		)

		curfewEnd := time.Date(
			currentDate.Year(), currentDate.Month(), currentDate.Day(),
			curfewEndHour, curfewEndMinute, 0, 0,
			currentDate.Location(),
		)

		// Handle overnight curfews
		if curfewEndHour < curfewStartHour || (curfewEndHour == curfewStartHour && curfewEndMinute < curfewStartMinute) {
			curfewEnd = curfewEnd.AddDate(0, 0, 1)
		}

		if !curfewStart.After(endTime) && !curfewEnd.Before(startTime) {
			windows = append(windows, TimeWindow{Start: curfewStart, End: curfewEnd})
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return windows
}

// findOptimalWindow finds the best time to schedule maintenance based on constraints.
func (p *IntelligentMaintenancePolicy) findOptimalWindow(
	preferredStart time.Time,
	endTime time.Time,
	curfewWindows []TimeWindow,
	existingMaintenance []maintenanceWindow,
) time.Time {
	duration := p.schedule.Duration

	// Try 1: During curfew (if maintenance fits entirely within curfew)
	for _, curfew := range curfewWindows {
		if curfew.Start.After(preferredStart) || curfew.Start.Equal(preferredStart) {
			if curfew.End.Sub(curfew.Start) >= duration {
				// Check runway coordination
				if p.checkRunwayCoordination(curfew.Start, curfew.Start.Add(duration), existingMaintenance) {
					return curfew.Start
				}
			}
		}
	}

	// Try 2: Adjacent to curfew start (maintenance ends when curfew starts)
	for _, curfew := range curfewWindows {
		adjacentStart := curfew.Start.Add(-duration)
		if !adjacentStart.Before(preferredStart) && adjacentStart.Add(duration).Before(endTime) {
			if p.checkRunwayCoordination(adjacentStart, adjacentStart.Add(duration), existingMaintenance) {
				return adjacentStart
			}
		}
	}

	// Try 3: Adjacent to curfew end (maintenance starts when curfew ends)
	for _, curfew := range curfewWindows {
		if !curfew.End.Before(preferredStart) && curfew.End.Add(duration).Before(endTime) {
			if p.checkRunwayCoordination(curfew.End, curfew.End.Add(duration), existingMaintenance) {
				return curfew.End
			}
		}
	}

	// Try 4: Fallback to preferred start if coordination allows
	if p.checkRunwayCoordination(preferredStart, preferredStart.Add(duration), existingMaintenance) {
		return preferredStart
	}

	// If all else fails, return zero time (caller will use current time)
	return time.Time{}
}

// checkRunwayCoordination ensures minimum operational runways are maintained.
func (p *IntelligentMaintenancePolicy) checkRunwayCoordination(
	proposedStart, proposedEnd time.Time,
	existingMaintenance []maintenanceWindow,
) bool {
	totalRunways := len(p.schedule.RunwayDesignations)

	// If we only have one runway, we must allow maintenance
	if totalRunways == 1 {
		return true
	}

	// Count how many runways would be in maintenance during this window
	concurrentMaintenance := 0

	for _, maint := range existingMaintenance {
		// Check if windows overlap
		if proposedStart.Before(maint.End) && proposedEnd.After(maint.Start) {
			concurrentMaintenance++
		}
	}

	// Check if we'd exceed the maximum concurrent maintenance
	maxConcurrentMaintenance := totalRunways - p.schedule.MinimumOperationalRunways

	return concurrentMaintenance < maxConcurrentMaintenance
}

// Helper to sort maintenance windows by start time
type byStartTime []TimeWindow

func (a byStartTime) Len() int           { return len(a) }
func (a byStartTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byStartTime) Less(i, j int) bool { return a[i].Start.Before(a[j].Start) }

func sortTimeWindows(windows []TimeWindow) {
	sort.Sort(byStartTime(windows))
}
