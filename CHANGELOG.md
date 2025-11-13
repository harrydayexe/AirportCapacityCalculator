# Changelog

All notable changes to the Airport Capacity Calculator project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
