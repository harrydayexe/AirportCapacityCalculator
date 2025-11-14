# Changelog

All notable changes to the Airport Capacity Calculator project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.5.0] - 2025-01-14

### Added

**Dynamic Wind System**
- `ScheduledWindPolicy` for time-varying wind conditions
- `WindChangeEvent` for modeling wind changes at specific times
- `policy.WindChange` struct for defining wind schedule entries
- `simulation.AddScheduledWindPolicy(schedule)` convenience method
- Automatic event generation within simulation time period
- Chronological validation of wind schedules

**Wind Pattern Helpers**
- `policy.DiurnalWindPattern()` for daily wind cycles (morning calm → afternoon build → evening decrease)
- `policy.ConstantWindPattern()` for simple constant wind conditions
- `policy.FrontalPassagePattern()` for abrupt wind shifts (weather fronts)
- `policy.LinearWindTransition()` for smooth wind transitions over time
- `policy.SeasonalWindPattern()` for seasonal wind variations
- `policy.CombineWindSchedules()` for merging multiple wind patterns
- `policy.SortSchedule()` utility for chronological sorting

**Event System Enhancements**
- `WindChangeType` event type added to event system
- `WorldState.SetWind()` interface method for wind state changes
- `WorldState.GetWindSpeed()` and `GetWindDirection()` accessors

### Changed

**Wind Policy Architecture**
- Wind changes now trigger `World.SetWind()` which automatically notifies `RunwayManager`
- RunwayManager recalculates active configuration on wind changes
- Event-driven wind updates maintain architectural consistency

### Usage Examples

**Scheduled Wind Policy:**
```go
// Create a wind schedule
schedule := []simulation.WindChange{
    {time.Date(2024, 1, 1, 6, 0, 0, 0, time.UTC), 5, 90},   // Morning: 5kt easterly
    {time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 20, 270}, // Afternoon: 20kt westerly
    {time.Date(2024, 1, 1, 21, 0, 0, 0, time.UTC), 10, 270}, // Evening: 10kt westerly
}

sim, err := simulation.NewSimulation(airport, logger).
    AddScheduledWindPolicy(schedule)
```

**Diurnal Wind Pattern:**
```go
import "github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/policy"

// 7-day diurnal pattern: morning calm → afternoon peak → evening decrease
schedule := policy.DiurnalWindPattern(
    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
    7,    // 7 days
    5.0,  // morning: 5kt
    20.0, // afternoon peak: 20kt
    10.0, // evening: 10kt
    270,  // westerly direction
)

sim, err := simulation.NewSimulation(airport, logger).
    AddScheduledWindPolicy(schedule)
```

**Frontal Passage:**
```go
// Model a cold front passage with abrupt wind shift
frontPassage := policy.FrontalPassagePattern(
    time.Date(2024, 3, 15, 18, 0, 0, 0, time.UTC),
    10,  // pre-frontal: 10kt southerly
    180,
    25,  // post-frontal: 25kt westerly
    270,
)

sim, err := simulation.NewSimulation(airport, logger).
    AddScheduledWindPolicy(frontPassage)
```

**Seasonal Wind Variation:**
```go
// Model seasonal prevailing winds
seasonal := policy.SeasonalWindPattern(
    2024,
    time.UTC,
    15, 10, 5, 12,      // speeds: winter, spring, summer, fall (knots)
    270, 180, 90, 225,  // directions: W, S, E, SW (degrees true)
)

sim, err := simulation.NewSimulation(airport, logger).
    AddScheduledWindPolicy(seasonal)
```

**Linear Wind Transition:**
```go
// Gradual wind change over 4 hours
transition, err := policy.LinearWindTransition(
    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
    4*time.Hour, // duration
    5,           // steps for smoothness
    10, 90,      // initial: 10kt from east
    30, 180,     // final: 30kt from south
)

sim, err := simulation.NewSimulation(airport, logger).
    AddScheduledWindPolicy(transition)
```

**Combining Wind Patterns:**
```go
// Combine seasonal and diurnal patterns
seasonal := policy.SeasonalWindPattern(2024, time.UTC, 15, 10, 5, 12, 270, 180, 90, 225)
diurnal := policy.DiurnalWindPattern(startDate, 365, 5, 20, 10, 270)
combined := policy.CombineWindSchedules(seasonal, diurnal)

sim, err := simulation.NewSimulation(airport, logger).
    AddScheduledWindPolicy(combined)
```

### Technical Details

**Wind Schedule Validation:**
- Schedule cannot be empty
- Wind speeds must be non-negative
- Wind directions are automatically normalized to 0-360° range
- Schedule entries must be in chronological order
- Duplicate timestamps: last entry wins

**Event Generation:**
- Only events within simulation period (startTime to endTime) are generated
- Events outside the period are silently skipped
- WindChangeEvents are processed in chronological order
- Each wind change triggers automatic runway configuration recalculation

**Direction Normalization:**
- Handles negative angles: -90° → 270°
- Handles angles >360°: 450° → 90°
- Ensures all directions are in 0-360° range

**Wind Pattern Helpers:**
- `DiurnalWindPattern`: 4 changes per day (midnight, 06:00, 15:00, 21:00)
- `LinearWindTransition`: Takes shortest angular path for direction changes
  - Example: 350° → 10° goes through 360°, not backwards through 180°
- `SeasonalWindPattern`: Changes at equinoxes/solstices (Mar 20, Jun 21, Sep 22, Jan 1)
- `CombineWindSchedules`: Merges and sorts multiple schedules chronologically

### Migration from Static Wind

**Before (v0.4.0 - Static Wind):**
```go
sim, err := simulation.NewSimulation(airport, logger).
    AddWindPolicy(15, 270) // Constant 15kt westerly
```

**After (v0.5.0 - Dynamic Wind):**
```go
// For constant wind throughout simulation
schedule := policy.ConstantWindPattern(
    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
    15, 270,
)
sim, err := simulation.NewSimulation(airport, logger).
    AddScheduledWindPolicy(schedule)

// Or use static WindPolicy (still supported)
sim, err := simulation.NewSimulation(airport, logger).
    AddWindPolicy(15, 270)
```

**Note:** Static `WindPolicy` from v0.4.0 remains fully supported for backward compatibility.

## [0.4.0] - 2025-01-13

### Added

**Wind-Based Runway Selection**
- `WindPolicy` for modeling wind conditions affecting runway usability
- Wind component calculations (headwind/crosswind decomposition)
- Automatic runway direction selection (Forward/Reverse) based on wind
- Runway filtering by crosswind and tailwind limits
- `Runway.CrosswindLimitKnots` field for maximum crosswind component
- `Runway.TailwindLimitKnots` field for maximum tailwind component
- `World.WindSpeed` and `World.WindDirection` for tracking wind state
- `RunwayManager.OnWindChanged()` for wind change notifications
- `simulation.AddWindPolicy(speed, direction)` convenience method

### Changed

**Enhanced Runway Manager**
- Modified `calculateActiveConfiguration()` to filter runways by wind constraints
- Added wind-based direction determination (prefers maximum headwind)
- RunwayManager now tracks wind state internally
- Wind changes trigger automatic runway configuration recalculation

**World State Management**
- `World.SetWind()` now notifies RunwayManager of wind changes
- Initial wind state defaults to calm (0kt, 0°)

### Usage Example

**Wind Policy:**
```go
// Add wind policy to simulation
sim, err := simulation.NewSimulation(airport, logger).
    AddWindPolicy(15, 270) // 15kt westerly wind
if err != nil {
    panic(err)
}

// Define runways with wind limits
runway := airport.Runway{
    RunwayDesignation:   "09L",
    TrueBearing:         90.0,
    CrosswindLimitKnots: 35.0, // Max crosswind
    TailwindLimitKnots:  10.0, // Max tailwind
    MinimumSeparation:   60 * time.Second,
}
```

**How Wind Affects Operations:**
- Runways exceeding crosswind limits are removed from operation
- Runways exceeding tailwind limits operate in reverse direction (if possible)
- Direction selection prefers maximum headwind component
- Calm wind (0kt) allows all runways in forward direction

**Example Scenarios:**
```go
// Westerly wind (270°) - Runways 09L/09R operate as 27R/27L (reverse)
AddWindPolicy(15, 270)

// Strong crosswind - May exclude some runways entirely
AddWindPolicy(40, 360)

// Calm conditions - All runways forward direction
AddWindPolicy(0, 0)
```

### Technical Details

**Wind Component Calculation:**
```
Given runway bearing and wind direction/speed:
- Headwind = speed × cos(angle_difference)
- Crosswind = speed × |sin(angle_difference)|
- Positive headwind = favorable (headwind)
- Negative headwind = unfavorable (tailwind)
```

**Runway Usability Logic:**
1. Calculate wind components for both directions (forward & reverse)
2. Check if either direction meets crosswind/tailwind limits
3. If both directions unusable → exclude runway entirely
4. If one direction usable → select that direction
5. If both usable → prefer direction with maximum headwind

## [0.3.1] - Unreleased

### Added

**Runway Compatibility System**
- `Airport.RunwayCompatibility` field for modeling crossing/parallel runway relationships
- `airport.NewRunwayCompatibility()` constructor for compatibility graph
- Automatic selection of maximum-capacity runway configuration

**Time-Bounded Rotation Policies**
- `RotationSchedule` struct for hour-of-day and day-of-week rotation windows
- `policy.NewRunwayRotationPolicyWithSchedule()` constructor

### Changed

**BREAKING: Rotation Strategy Renamed**
- `simulation.BalancedRotation` → `simulation.PreferentialRunway`

**Migration:**
```go
// Before
sim.RunwayRotationPolicy(simulation.BalancedRotation)

// After
sim.RunwayRotationPolicy(simulation.PreferentialRunway)
```

### Usage Example

**Runway Compatibility:**
```go
airport := airport.Airport{
    Name: "Example Airport",
    Runways: []airport.Runway{ /* runways */ },
    RunwayCompatibility: airport.NewRunwayCompatibility(map[string][]string{
        "09L": {"09R"},  // Parallel runways (compatible)
        "09R": {"09L"},
        "18":  {},       // Crosses both (incompatible)
    }),
}
```

**Time-Bounded Rotation:**
```go
schedule := &simulation.RotationSchedule{
    StartHour:  6,   // 6 AM
    EndHour:    23,  // 11 PM
    DaysOfWeek: []time.Weekday{time.Saturday, time.Sunday},
}
policy := policy.NewRunwayRotationPolicyWithSchedule(
    simulation.TimeBasedRotation,
    policy.NewDefaultRotationPolicyConfiguration(),
    schedule,
)
```

---

## [0.3.0] - 2024-11-01

### Added

**Intelligent Maintenance Policy**
- `AddIntelligentMaintenancePolicy()` - curfew-aware maintenance scheduling
- Ensures minimum operational runways during maintenance

**Gate Capacity Policy**
- `AddGateCapacityPolicy()` - models gate capacity constraints
- Automatically caps runway capacity when gates are the bottleneck

**Taxi Time Policy**
- `AddTaxiTimePolicy()` - models taxi time impact on gate capacity

### API Example

```go
sim, err := simulation.NewSimulation(airport, logger).
    AddGateCapacityPolicy(simulation.GateCapacityConstraint{
        TotalGates:            50,
        AverageTurnaroundTime: 2 * time.Hour,
    }).
    AddTaxiTimePolicy(simulation.TaxiTimeConfiguration{
        AverageTaxiInTime:  5 * time.Minute,
        AverageTaxiOutTime: 5 * time.Minute,
    })
```

---

## [0.2.0] - 2024-10-31

### Changed

**BREAKING: Event-Driven Architecture**
- Policies now use `GenerateEvents()` instead of `Apply()`
- `Airport.MinimumSeparation` field removed (use per-runway `Runway.MinimumSeparation`)
- `SimulationState` removed (replaced with `World`)

### Migration Guide

**Airport Configuration:**
```go
// Before (v0.1.0)
airport := Airport{
    MinimumSeparation: 60 * time.Second,
    Runways: []Runway{{...}},
}

// After (v0.2.0+)
airport := Airport{
    Runways: []Runway{{
        MinimumSeparation: 60 * time.Second,
        ...
    }},
}
```

**Custom Policy Implementation:**
```go
// Before (v0.1.0)
type MyPolicy struct{}
func (p *MyPolicy) Apply(ctx context.Context, state any, logger *slog.Logger) error {
    // Modify state directly
}

// After (v0.2.0+)
type MyPolicy struct{}
func (p *MyPolicy) GenerateEvents(ctx context.Context, world EventWorld) error {
    world.ScheduleEvent(myEvent)
    return nil
}
```

---

## [0.1.0] - Initial Release

- Basic airport capacity calculation
- Policy system: Curfew, Maintenance, Rotation
- Formula-based aggregate calculations
