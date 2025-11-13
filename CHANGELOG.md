# Changelog

All notable changes to the Airport Capacity Calculator project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
