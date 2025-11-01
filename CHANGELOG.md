# Changelog

All notable changes to the Airport Capacity Calculator project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - 2024-11-01

### Added - Intelligence Layer

**Intelligent Maintenance Scheduling** (`internal/simulation/policy/intelligent_maintenance.go`)
- Curfew-aware maintenance scheduling (prefers maintenance during operational downtime)
- Peak hours avoidance configuration
- Runway coordination (ensures minimum runways remain operational)
- Automatic staggering of maintenance across runway fleet
- Reduces capacity impact by optimizing when maintenance occurs

**Gate Capacity Constraints** (`internal/simulation/policy/gate_capacity.go`)
- Models limitation that gates place on sustained throughput
- Calculates maximum sustainable movements based on gates and turnaround time
- Formula: `sustained_capacity = (gates / turnaround_time) × 2`
- Automatic capacity capping when gates are more restrictive than runways
- Integrates with event-driven engine for accurate bottleneck modeling

**Taxi Time Modeling** (`internal/simulation/policy/taxi_time.go`)
- Models impact of taxi time on sustained capacity
- Configurable taxi-in and taxi-out times
- Adjusts effective gate capacity based on taxi overhead
- Accounts for additional time aircraft spend at airport beyond gate occupancy
- Formula: `effective_capacity = gates / (turnaround + taxi_overhead)`

### Changed

**Engine Capacity Calculation**
- Added gate capacity constraint handling in `calculateWindowCapacity()`
- Taxi time overhead integrated with gate capacity calculations
- Capacity now respects multiple bottlenecks (runways, gates, taxi time)

**World State**
- Added `GateCapacityConstraint` field (max movements/second limited by gates)
- Added `TaxiTimeOverhead` field (total taxi time per aircraft cycle)
- Added getter/setter methods for new state fields

**Event System**
- Added `GateCapacityConstraintType` event type
- Added `TaxiTimeAdjustmentType` event type
- Added `GateCapacityConstraintEvent` implementation
- Added `TaxiTimeAdjustmentEvent` implementation

### Fixed

**Documentation**
- Corrected outdated documentation in CLAUDE.md that incorrectly stated `RunwayDesigntation` field had a typo (field was already correctly named `RunwayDesignation`)
- Added `MinimumSeparation` field to Runway documentation in CLAUDE.md
- Added project binary to `.gitignore` to prevent build artifacts in repository

### Technical Details

**Intelligent Maintenance Scheduling Algorithm:**
1. Try to schedule during curfew (if maintenance fits entirely within curfew)
2. Try to schedule adjacent to curfew start (maintenance ends when curfew starts)
3. Try to schedule adjacent to curfew end (maintenance starts when curfew ends)
4. Avoid peak hours if configured
5. Ensure minimum operational runways maintained at all times
6. Stagger maintenance across runways (offset start times)

**Gate Capacity Impact:**
- If runways can handle 100 movements/hour but gates only support 60 movements/hour, capacity is capped at 60
- Taxi time overhead reduces effective gate capacity by extending aircraft dwell time

**Performance:**
- All new policies use event generation pattern (no runtime state modification)
- Minimal performance overhead from additional constraint checks
- Capacity calculations remain O(n) where n = number of state-changing events

### Migration Notes

**Using New Policies:**
```go
// Intelligent maintenance scheduling
sim, err := simulation.NewSimulation(airport, logger).
    AddIntelligentMaintenancePolicy(simulation.IntelligentMaintenanceSchedule{
        RunwayDesignations:       []string{"09L", "09R"},
        Duration:                 4 * time.Hour,
        Frequency:                30 * 24 * time.Hour,
        MinimumOperationalRunways: 1,
        CurfewStart:              &curfewStart,
        CurfewEnd:                &curfewEnd,
        PeakHours:                &simulation.PeakHours{StartHour: 6, EndHour: 22},
    })

// Gate capacity constraints
sim, err = sim.AddGateCapacityPolicy(simulation.GateCapacityConstraint{
    TotalGates:            50,
    AverageTurnaroundTime: 2 * time.Hour,
})

// Taxi time modeling
sim, err = sim.AddTaxiTimePolicy(simulation.TaxiTimeConfiguration{
    AverageTaxiInTime:  5 * time.Minute,
    AverageTaxiOutTime: 5 * time.Minute,
})
```

## [0.2.0] - 2024-10-31

### Major: Event-Driven Simulation Engine Redesign

Complete architectural redesign from formula-based aggregate calculations to event-driven state-window simulation.

#### Added

**Event System** (`internal/simulation/event/`)
- `Event` interface for time-based state changes
- `EventQueue` priority queue for chronological event processing
- `CurfewStartEvent` and `CurfewEndEvent` for operation restrictions
- `RunwayMaintenanceStartEvent` and `RunwayMaintenanceEndEvent` for runway availability changes
- `RotationChangeEvent` for efficiency multiplier updates
- `EventType` enum with string representations
- `WorldState` interface for event-world interaction

**World State Management** (`internal/simulation/world.go`)
- `World` struct tracking complete simulation state
- `RunwayState` for per-runway availability tracking
- `NewWorld()` constructor with time boundaries and runway initialization
- State accessor methods: `GetAvailableRunways()`, `CountAvailableRunways()`
- `EventWorld` interface implementation for policy interaction

**Policy System Redesign**
- New `Policy` interface with `GenerateEvents()` method (replaces `Apply()`)
- `EventWorld` interface in policy package to avoid circular dependencies
- `EventGeneratingPolicy` interface for new policy architecture
- Test helpers: `mockEventWorld` for policy testing

**Engine Enhancements** (`internal/simulation/engine.go`)
- `Calculate()` method with event-driven timeline processing
- `processTimeline()` for chronological event processing
- `calculateWindowCapacity()` for discrete time window calculations
- Per-runway capacity calculation using individual separation times
- Rotation multiplier application to window capacity

**Simulation Orchestration** (`internal/simulation/simulation.go`)
- `Run()` method fully migrated to event-driven approach
- Automatic simulation time boundaries (one year)
- Event generation from all policies before processing
- Comprehensive logging for event processing

**Testing Infrastructure**
- `test_helpers.go` with `mockEventWorld` and `testLogger()`
- Event counting and verification helpers
- Comprehensive test suite for all policies using event-driven approach

#### Changed

**Breaking Changes**

**Airport Model** (`internal/airport/airport.go`)
- **BREAKING**: Removed `MinimumSeparation` field from `Airport` struct
- **Rationale**: Enables per-runway separation configuration
- **Migration**: Move separation to individual `Runway` structs

**Runway Model** (`internal/airport/runway.go`)
- `MinimumSeparation` field already existed per-runway
- Now the only source of separation configuration

**Policy Interfaces**
- **BREAKING**: Removed `Apply(ctx, state, logger)` method from all policies
- **BREAKING**: Removed support for direct state modification
- **Migration**: Policies must implement `GenerateEvents(ctx, world)` instead

**CurfewPolicy** (`internal/simulation/policy/curfew.go`)
- Complete rewrite to generate daily curfew events
- `GenerateEvents()` creates paired start/end events for each day
- Overnight curfew support (e.g., 11 PM to 6 AM next day)
- Events generated for entire simulation period (365+ days)
- Removed legacy `Apply()` method
- Removed unused imports (`fmt`, `log/slog`)

**MaintenancePolicy** (`internal/simulation/policy/maintenance.go`)
- Complete rewrite to generate per-runway maintenance events
- `GenerateEvents()` schedules maintenance windows based on frequency
- Maintenance windows distributed evenly across simulation period
- Per-runway validation against world's runway list
- Supports multiple runways with independent schedules
- Removed legacy `Apply()` method
- Removed unused imports

**RunwayRotationPolicy** (`internal/simulation/policy/rotation.go`)
- Complete rewrite to generate single rotation efficiency event
- `GenerateEvents()` schedules one `RotationChange` event at simulation start
- Efficiency multiplier set for entire simulation duration
- Custom configuration support maintained
- Removed legacy `Apply()` method
- Removed unused imports (`log/slog`)

**Simulation Engine** (`internal/simulation/engine.go`)
- Removed legacy `Calculate()` method that used `SimulationState`
- Removed `calculateOperatingSeconds()` (no longer needed)
- Removed old `calculateRunwayCapacity()` with aggregate calculation
- New `Calculate()` method signature: `Calculate(ctx, *World)` instead of `Calculate(ctx, *SimulationState)`
- Engine now processes events instead of modified state

**Simulation Orchestration** (`internal/simulation/simulation.go`)
- Removed `SimulationState` struct entirely
- Removed `runLegacy()` and `runWithEvents()` methods
- Removed backward compatibility routing logic
- Single `Run()` method using only event-driven approach
- Removed state helper methods: `GetOperatingHours()`, `SetOperatingHours()`, etc.

**Test Suite**
- Complete rewrite of all policy tests
- `curfew_test.go`: Event generation validation, overnight curfews, full year simulation
- `maintenance_test.go`: Multiple runways, frequency/duration verification, invalid runway handling
- `rotation_test.go`: Efficiency multiplier verification, custom configuration tests
- All tests migrated from state-checking to event-counting approach

#### Removed

**Legacy Components**
- `SimulationState` struct and all references
- Legacy `Policy.Apply()` interface method
- Backward compatibility code for old policy system
- `Engine.calculateOperatingSeconds()` method
- Old `Engine.calculateRunwayCapacity()` aggregate formula method
- `mockSimulationState` test helper (replaced with `mockEventWorld`)
- Global `Airport.MinimumSeparation` field

**Unused Code**
- Unused imports in policy files after migration
- Legacy test cases using `Apply()` method
- State modification helpers in `simulation.go`

#### Fixed

**Curfew Event Generation**
- Fixed boundary condition for events at simulation end time
- Changed from `.Before(endTime)` to `!.After(endTime)` for inclusive boundaries
- Ensures curfew end events at exact end time are included

**Engine Capacity Calculation**
- Fixed per-runway capacity calculation to use individual separation times
- Resolved issue where all runways assumed same separation
- Corrected formula: `Σ(duration / runway.MinimumSeparation) × rotationMultiplier`

**Test Expectations**
- Adjusted test simulation end times to include final curfew end events
- Fixed leap year test expectations (366 days vs 365)
- Corrected event count assertions for overnight curfews

#### Technical Details

**Architecture Improvements**
- **Separation of Concerns**: Policies generate events, engine calculates capacity
- **Time-Window Calculation**: Discrete windows between state changes
- **Chronological Processing**: Priority queue ensures correct event order
- **State Immutability**: World state only modified via events
- **Extensibility**: Easy to add new event types and policies

**Performance Characteristics**
- Event processing: O(n log n) where n = number of events
- Memory usage: O(n) for event queue
- Typical event count: 732 for full year (1 rotation + 731 curfew)
- Simulation time: Sub-second for annual calculations

**Capacity Formula**

For each time window [t₁, t₂]:
```
window_capacity = Σ(duration / runway.separation) × rotation_multiplier

If curfew_active:
    window_capacity = 0

total_capacity = Σ(all windows)
```

**Event Processing Flow**
```
1. Policies generate events → EventQueue
2. Sort events by time (priority queue)
3. For each event:
   a. Calculate capacity [previous_time, event_time]
   b. Apply event (modify world state)
   c. Update previous_time
4. Calculate final window [last_event, end_time]
5. Return total capacity
```

#### Migration Guide

**For Existing Code Using Old API:**

1. **Update Airport Configuration:**
```go
// OLD
airport := Airport{
    MinimumSeparation: 60 * time.Second,
    Runways: []Runway{{...}},
}

// NEW
airport := Airport{
    Runways: []Runway{{
        MinimumSeparation: 60 * time.Second,
        ...
    }},
}
```

2. **Update Policy Usage:**
```go
// OLD - Not supported anymore
type MyPolicy struct{}
func (p *MyPolicy) Apply(ctx, state, logger) error { ... }

// NEW
type MyPolicy struct{}
func (p *MyPolicy) GenerateEvents(ctx, world EventWorld) error {
    world.ScheduleEvent(myEvent)
    return nil
}
```

3. **Update Tests:**
```go
// OLD
state := &mockSimulationState{...}
policy.Apply(ctx, state, logger)
assert(state.operatingHours == expected)

// NEW
world := newMockEventWorld(start, end, runways)
policy.GenerateEvents(ctx, world)
assert(world.CountEventsByType(EventType) == expected)
```

#### Known Limitations

**Current Implementation:**
- Maintenance windows distributed evenly (not optimized)
- Rotation efficiency constant throughout simulation
- No support for crossing runway conflicts
- No gate capacity constraints
- No taxi time modeling

**Planned Enhancements:**
- Intelligent maintenance scheduling (coordinate with rotation)
- Dynamic rotation efficiency (time-based changes)
- Gate capacity constraints
- Crossing runway conflict detection
- Taxi time and runway occupation modeling

#### Documentation Updates

**New Documentation:**
- Comprehensive README with event-driven architecture
- Policy usage examples and configuration
- Development guide for adding policies/events
- Test suite documentation
- Performance characteristics
- MIT License added (LICENSE file)

**Updated Documentation:**
- CLAUDE.md with event-driven policy guidelines
- README with current architecture and examples
- Inline code documentation
- Test file documentation

#### Dependencies

**No New Dependencies Added**
- Pure Go standard library implementation
- `container/heap` for priority queue
- `log/slog` for structured logging
- `time` for temporal operations
- `context` for cancellation support

#### Backward Compatibility

**Breaking Changes - No Backward Compatibility**
- Complete redesign requires migration
- Old policy API removed entirely
- State-based approach no longer supported
- Tests must be rewritten for event-driven approach

**Rationale:**
- Clean architecture without legacy code
- Simpler codebase to maintain
- Foundation for future enhancements
- Better separation of concerns

---

## Version History

### [0.2.0] - 2024-10-31

**Event-Driven Simulation Engine Release**

Major architectural redesign as documented above.

### [0.1.0] - Previous

**Initial Formula-Based Implementation**

- Basic airport capacity calculation
- Formula-based approach with aggregate state
- Simple policy system with direct state modification
- Single-runway support with global separation

---

## Future Releases

### Planned for [0.3.0]

**Intelligent Maintenance Scheduling**
- Coordinate maintenance with rotation periods
- Optimize for minimal capacity impact
- `MaintenanceOptimizer` implementation

**Enhanced Testing**
- Integration tests for complex scenarios
- Performance benchmarks
- Validation against real-world data

### Planned for [0.4.0]

**Gate Capacity Constraints**
- Gate state tracking
- Arrival/departure gate requirements
- Gate-limited capacity calculations

**Crossing Runway Support**
- Runway compatibility matrix
- Conflict detection and resolution
- Capacity adjustments for crossing operations

### Future Considerations

- Weather impact modeling
- Real-time optimization
- Multi-airport simulations
- Historical data integration
- Visualization tools
- REST API for simulations

---

## Notes

**Testing Coverage:**
- All policy tests passing
- Event generation validated
- Edge cases covered (overnight curfews, boundaries, etc.)
- Integration with full simulation verified

**Performance Verified:**
- Sub-second simulation for full year
- Minimal memory footprint
- Efficient event processing
- No performance regressions

**Code Quality:**
- All tests passing: `go test ./...`
- No unused code or imports
- Comprehensive documentation
- Clean architecture with clear separation
