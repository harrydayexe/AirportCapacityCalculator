# Airport Capacity Calculator

A data experiment to calculate the theoretical maximum capacity of airport movements using event-driven simulation. Models real-world operational constraints including curfews, maintenance windows, and runway rotation strategies.

## Overview

The Airport Capacity Calculator uses an **event-driven simulation engine** to calculate theoretical maximum aircraft movements at airports. The simulator processes chronological events (curfews, maintenance, rotation changes) and calculates capacity for discrete time windows throughout the year.

### Key Features

- **Event-Driven Architecture**: Time-based state changes processed chronologically
- **Policy System**: Modular, reusable policies for operational constraints
- **Per-Runway Configuration**: Individual separation times and maintenance schedules
- **Runway Rotation Strategies**: Model efficiency impacts of different rotation approaches
- **Curfew Modeling**: Overnight and multi-day curfew support
- **Maintenance Scheduling**: Per-runway maintenance with configurable frequency and duration

## Getting Started

### Prerequisites

- Go 1.24.4 or higher

### Installation

```bash
# Clone the repository
git clone https://github.com/harrydayexe/AirportCapacityCalculator.git
cd AirportCapacityCalculator

# Build the application
go build -o airportCapacityCalculator ./cmd/airportCapacityCalculator.go

# Or run directly
go run ./cmd/airportCapacityCalculator.go
```

### Quick Start Example

```go
package main

import (
    "context"
    "log/slog"
    "os"
    "time"

    "github.com/harrydayexe/AirportCapacityCalculator/internal/airport"
    "github.com/harrydayexe/AirportCapacityCalculator/internal/simulation"
)

func main() {
    // Define an airport
    myAirport := airport.Airport{
        Name:     "Example International",
        IATACode: "EXA",
        ICAOCode: "KEXA",
        City:     "Example City",
        Country:  "Example Country",
        Runways: []airport.Runway{
            {
                RunwayDesignation: "09L",
                TrueBearing:       90.0,
                LengthMeters:      3000.0,
                WidthMeters:       45.0,
                SurfaceType:       airport.Asphalt,
                MinimumSeparation: 60 * time.Second,
            },
        },
    }

    logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

    // Create simulation with policies
    curfewStart := time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC)
    curfewEnd := time.Date(2024, 1, 2, 6, 0, 0, 0, time.UTC)

    sim, err := simulation.NewSimulation(myAirport, logger).
        AddCurfewPolicy(curfewStart, curfewEnd)
    if err != nil {
        panic(err)
    }

    sim = sim.RunwayRotationPolicy(simulation.NoRotation)

    // Run simulation
    capacity, err := sim.Run(context.Background())
    if err != nil {
        panic(err)
    }

    logger.Info("Annual capacity calculated", "movements", int(capacity))
}
```

## Architecture

### Event-Driven Simulation

The simulation engine processes events chronologically to calculate capacity:

```
Policies → Generate Events → Priority Queue → Process Timeline → Calculate Capacity
```

**Key Components:**

1. **Events** (`internal/simulation/event/`): State changes at specific times
   - `CurfewStart/End`: Operations stop/resume
   - `RunwayMaintenanceStart/End`: Runway availability changes
   - `RotationChange`: Efficiency multiplier changes

2. **World** (`internal/simulation/world.go`): Current simulation state
   - Runway availability
   - Curfew status
   - Rotation efficiency

3. **Policies** (`internal/simulation/policy/`): Generate events based on operational rules
   - `CurfewPolicy`: Restricts operations during specified hours
   - `MaintenancePolicy`: Schedules runway maintenance
   - `RunwayRotationPolicy`: Applies efficiency multipliers

4. **Engine** (`internal/simulation/engine.go`): Processes events and calculates capacity
   - Processes events chronologically
   - Calculates capacity for time windows
   - Aggregates annual capacity

### Project Structure

```
.
├── cmd/
│   └── airportCapacityCalculator.go    # Main application demonstrating rotation strategies
├── internal/
│   ├── airport/
│   │   ├── airport.go                  # Airport model
│   │   └── runway.go                   # Runway model with operational parameters
│   └── simulation/
│       ├── simulation.go               # Simulation orchestrator
│       ├── engine.go                   # Event processing and capacity calculation
│       ├── world.go                    # World state management
│       ├── event/
│       │   ├── event.go               # Event interface and types
│       │   ├── queue.go               # Priority queue for time-ordered events
│       │   ├── curfew.go              # Curfew events
│       │   ├── maintenance.go         # Maintenance events
│       │   └── rotation.go            # Rotation events
│       └── policy/
│           ├── curfew.go              # Curfew policy implementation
│           ├── maintenance.go         # Maintenance policy implementation
│           ├── rotation.go            # Rotation policy implementation
│           └── constants.go           # Shared constants
├── testdata/                           # Test fixtures
├── CLAUDE.md                           # Development guide and policy documentation
├── CHANGELOG.md                        # Detailed changelog
└── README.md                           # This file
```

## Policies

### Curfew Policy

Restricts airport operations during specified hours.

```go
curfewStart := time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC)  // 11 PM
curfewEnd := time.Date(2024, 1, 2, 6, 0, 0, 0, time.UTC)     // 6 AM next day

sim, err := simulation.NewSimulation(airport, logger).
    AddCurfewPolicy(curfewStart, curfewEnd)
```

**Features:**
- Overnight curfew support (e.g., 11 PM - 6 AM)
- Daily event generation for full simulation period
- Validation (max 30-day duration)

### Maintenance Policy

Schedules recurring maintenance windows for specific runways.

```go
schedule := simulation.MaintenanceSchedule{
    RunwayDesignations: []string{"09L", "27R"},
    Duration:           4 * time.Hour,
    Frequency:          30 * 24 * time.Hour,  // Monthly
}

sim := simulation.NewSimulation(airport, logger).
    AddMaintenancePolicy(schedule)
```

**Features:**
- Per-runway maintenance windows
- Configurable frequency and duration
- Distributed evenly across simulation period
- Runway availability validation

### Runway Rotation Policy

Models efficiency impacts of runway rotation strategies.

```go
sim := simulation.NewSimulation(airport, logger).
    RunwayRotationPolicy(simulation.TimeBasedRotation)
```

**Strategies:**

| Strategy | Efficiency | Description |
|----------|-----------|-------------|
| `NoRotation` | 100% | Maximum efficiency, runways used optimally |
| `TimeBasedRotation` | 95% | 5% overhead for transition periods |
| `BalancedRotation` | 90% | 10% penalty for distributing usage evenly |
| `NoiseOptimizedRotation` | 80% | 20% reduction for noise mitigation |

**Custom Configuration:**
```go
customConfig := policy.NewRotationPolicyConfiguration(map[RotationStrategy]float32{
    NoRotation:             0.99,
    TimeBasedRotation:      0.85,
    BalancedRotation:       0.75,
    NoiseOptimizedRotation: 0.65,
})

policy := policy.NewRunwayRotationPolicy(BalancedRotation, customConfig)
sim := simulation.NewSimulation(airport, logger).AddPolicy(policy)
```

## Running Simulations

### Example Simulation

The included example (`cmd/airportCapacityCalculator.go`) demonstrates rotation strategies:

```bash
go run ./cmd/airportCapacityCalculator.go
```

**Output:**
```
Scenario 1: No Rotation (baseline - maximum efficiency)
Result capacity=1121040 strategy=NoRotation

Scenario 2: Time-Based Rotation (5% capacity reduction)
Result capacity=1064988 strategy=TimeBasedRotation

Scenario 3: Balanced Rotation (10% capacity reduction)
Result capacity=1008936 strategy=BalancedRotation

Scenario 4: Noise-Optimized Rotation (20% capacity reduction)
Result capacity=896832 strategy=NoiseOptimizedRotation

=== Capacity Comparison ===
NoRotation Capacity: 1121040
TimeBasedRotation Capacity: 1064988
BalancedRotation Capacity: 1008936
NoiseOptimizedRotation Capacity: 896832

=== Impact Analysis ===
TimeBasedRotation Reduction: 56052
BalancedRotation Reduction: 112104
NoiseOptimizedRotation Reduction: 224208
```

### Custom Simulations

Create custom simulations by combining policies:

```go
// Complex scenario: 3 runways, curfew, maintenance, rotation
sim, err := simulation.NewSimulation(airport, logger).
    AddCurfewPolicy(curfewStart, curfewEnd).
    AddMaintenancePolicy(maintenanceSchedule).
    RunwayRotationPolicy(simulation.BalancedRotation)
if err != nil {
    panic(err)
}

capacity, err := sim.Run(context.Background())
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/simulation/policy

# Run with coverage
go test -cover ./...
```

### Test Coverage

- **Policy Tests**: Event generation, validation, edge cases
- **Event Tests**: Time ordering, state changes
- **Integration Tests**: Full simulation scenarios

## Development

### Adding a New Policy

1. Create policy struct in `internal/simulation/policy/`:
```go
type MyPolicy struct {
    // configuration fields
}

func (p *MyPolicy) Name() string {
    return "MyPolicy"
}

func (p *MyPolicy) GenerateEvents(ctx context.Context, world EventWorld) error {
    // Generate events based on policy logic
    world.ScheduleEvent(myEvent)
    return nil
}
```

2. Add convenience method to `Simulation`:
```go
func (s *Simulation) AddMyPolicy(config MyConfig) (*Simulation, error) {
    p := policy.NewMyPolicy(config)
    return s.AddPolicy(p), nil
}
```

3. Write tests in `internal/simulation/policy/mypolicy_test.go`

### Creating Custom Events

1. Define event in `internal/simulation/event/`:
```go
type MyEvent struct {
    timestamp time.Time
    data      string
}

func (e *MyEvent) Time() time.Time { return e.timestamp }
func (e *MyEvent) Type() EventType { return MyEventType }
func (e *MyEvent) Apply(ctx context.Context, world WorldState) error {
    // Modify world state
    return nil
}
```

2. Add event type to `EventType` enum
3. Update `EventType.String()` method

See `CLAUDE.md` for detailed development guidelines.

## Future Enhancements

The architecture supports these planned features:

- **Intelligent Maintenance Scheduling**: Optimize maintenance during low-usage periods
- **Gate Capacity Constraints**: Limit operations based on available gates
- **Taxi Time Modeling**: Track runway occupation beyond separation time
- **Crossing Runway Conflicts**: Model runway dependencies
- **Weather Impact**: Dynamic capacity adjustments
- **Real-time Optimization**: Dynamic event generation based on conditions

## Performance

**Typical Performance** (3-runway airport, full year simulation):
- Events processed: ~732 (1 rotation + 731 curfew events)
- Simulation time: <1 second
- Memory usage: Minimal (events processed sequentially)

## Contributing

Contributions welcome! Please:

1. Follow Go conventions and project structure
2. Add tests for new features
3. Update documentation (README, CLAUDE.md)
4. Ensure all tests pass: `go test ./...`

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for detailed version history.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

Developed with assistance from Claude Code (Anthropic).
