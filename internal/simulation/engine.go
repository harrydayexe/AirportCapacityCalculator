package simulation

import (
	"context"
	"log/slog"
	"time"
)

// Engine is the core event-driven simulation engine that calculates total movements
// by processing events chronologically and calculating capacity for each time window.
type Engine struct {
	logger *slog.Logger
}

// NewEngine creates a new simulation engine.
func NewEngine(logger *slog.Logger) *Engine {
	return &Engine{
		logger: logger,
	}
}

// Calculate computes total annual movements using event-driven state-window approach.
// This method processes events chronologically and calculates capacity for each time window.
func (e *Engine) Calculate(ctx context.Context, world *World) (float32, error) {
	e.logger.InfoContext(ctx, "Starting event-driven capacity calculation",
		"airport", world.Airport.Name,
		"startTime", world.StartTime,
		"endTime", world.EndTime,
		"numEvents", world.Events.Len())

	totalCapacity, err := e.processTimeline(ctx, world)
	if err != nil {
		return 0, err
	}

	e.logger.InfoContext(ctx, "Event-driven calculation complete", "totalCapacity", totalCapacity)

	return totalCapacity, nil
}

// processTimeline processes events chronologically and calculates capacity for each time window.
func (e *Engine) processTimeline(ctx context.Context, world *World) (float32, error) {
	totalCapacity := float32(0)
	previousEventTime := world.StartTime

	e.logger.InfoContext(ctx, "Processing timeline", "numEvents", world.Events.Len())

	// Process events in chronological order
	eventCount := 0
	for world.Events.HasNext() {
		evt := world.Events.Pop()
		eventTime := evt.Time()

		// Skip events outside simulation period
		if eventTime.Before(world.StartTime) {
			e.logger.DebugContext(ctx, "Skipping event before start time",
				"eventType", evt.Type().String(),
				"eventTime", eventTime,
				"startTime", world.StartTime)
			continue
		}

		if eventTime.After(world.EndTime) {
			e.logger.DebugContext(ctx, "Skipping event after end time",
				"eventType", evt.Type().String(),
				"eventTime", eventTime,
				"endTime", world.EndTime)
			// Put it back for final window calculation
			previousEventTime = world.EndTime
			break
		}

		// Calculate capacity for window [previousEventTime, eventTime]
		windowDuration := eventTime.Sub(previousEventTime)
		windowCapacity := e.calculateWindowCapacity(ctx, world, windowDuration)

		e.logger.DebugContext(ctx, "Window capacity calculated",
			"windowStart", previousEventTime,
			"windowEnd", eventTime,
			"duration", windowDuration,
			"capacity", windowCapacity)

		totalCapacity += windowCapacity

		// Apply event (changes world state)
		e.logger.InfoContext(ctx, "Applying event",
			"eventType", evt.Type().String(),
			"eventTime", eventTime)

		if err := evt.Apply(ctx, world); err != nil {
			e.logger.ErrorContext(ctx, "Failed to apply event",
				"eventType", evt.Type().String(),
				"error", err)
			return 0, err
		}

		world.CurrentTime = eventTime
		previousEventTime = eventTime
		eventCount++
	}

	// Calculate capacity for final window from last event to end of simulation
	if previousEventTime.Before(world.EndTime) {
		finalDuration := world.EndTime.Sub(previousEventTime)
		finalCapacity := e.calculateWindowCapacity(ctx, world, finalDuration)

		e.logger.DebugContext(ctx, "Final window capacity calculated",
			"windowStart", previousEventTime,
			"windowEnd", world.EndTime,
			"duration", finalDuration,
			"capacity", finalCapacity)

		totalCapacity += finalCapacity
	}

	e.logger.InfoContext(ctx, "Timeline processing complete",
		"eventsProcessed", eventCount,
		"totalCapacity", totalCapacity)

	return totalCapacity, nil
}

// calculateWindowCapacity calculates the theoretical maximum capacity for a time window
// given the current world state (runway availability, curfew, rotation multiplier).
func (e *Engine) calculateWindowCapacity(ctx context.Context, world *World, duration time.Duration) float32 {
	// If curfew is active, no operations are possible
	if world.CurfewActive {
		return 0
	}

	durationSeconds := float32(duration.Seconds())
	capacity := float32(0)

	// Sum capacity across all available runways
	for _, runwayState := range world.RunwayStates {
		if !runwayState.Available {
			continue
		}

		separationSeconds := float32(runwayState.Runway.MinimumSeparation.Seconds())

		// Runway capacity = duration / separation
		runwayCapacity := durationSeconds / separationSeconds
		capacity += runwayCapacity
	}

	// Apply rotation efficiency multiplier
	capacity *= world.RotationMultiplier

	// Apply gate capacity constraint if present
	if world.GateCapacityConstraint > 0 {
		// Gate constraint is in movements per second
		effectiveGateConstraint := world.GateCapacityConstraint

		// If taxi time overhead is configured, adjust gate capacity
		if world.TaxiTimeOverhead > 0 {
			// Taxi time extends the effective turnaround time, reducing sustainable capacity
			// For example: if base constraint allows 50 mvmt/hour (1 mvmt/72s)
			// and taxi adds 10 min (600s) overhead, effective becomes 1 mvmt/(72s+600s)

			// Calculate movements per second with taxi overhead
			// Original: 1 movement per X seconds
			// With taxi: 1 movement per (X + taxi_overhead) seconds
			baseSecondsPerMovement := float32(1.0) / effectiveGateConstraint
			taxiOverheadSeconds := float32(world.TaxiTimeOverhead.Seconds())
			adjustedSecondsPerMovement := baseSecondsPerMovement + taxiOverheadSeconds
			effectiveGateConstraint = 1.0 / adjustedSecondsPerMovement

			e.logger.DebugContext(ctx, "Taxi time overhead applied to gate capacity",
				"baseGateConstraint", world.GateCapacityConstraint,
				"effectiveGateConstraint", effectiveGateConstraint,
				"taxiOverhead", world.TaxiTimeOverhead)
		}

		// Convert to movements for this duration
		gateConstrainedCapacity := effectiveGateConstraint * durationSeconds

		// Take the minimum of runway capacity and gate-constrained capacity
		if gateConstrainedCapacity < capacity {
			e.logger.DebugContext(ctx, "Gate capacity constraint applied",
				"runwayCapacity", capacity,
				"gateConstrainedCapacity", gateConstrainedCapacity,
				"duration", duration)
			capacity = gateConstrainedCapacity
		}
	}

	return capacity
}
