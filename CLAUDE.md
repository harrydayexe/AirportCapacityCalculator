# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Airport Capacity Calculator is a Go-based data experiment for calculating the theoretical maximum capacity of airport movements. The project is in early development stages.

**Module Path:** `github.com/harrydayexe/AirportCapacityCalculator`
**Go Version:** 1.24.4

## Commands

### Building and Running
```bash
# Build the main binary
go build -o airportCapacityCalculator ./cmd/airportCapacityCalculator.go

# Run the application
go run ./cmd/airportCapacityCalculator.go
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test ./internal/airport/...
go test ./internal/airport/runway/...

# Run with coverage
go test -cover ./...
```

### Code Quality
```bash
# Format code
go fmt ./...

# Vet code for common issues
go vet ./...

# Run static analysis (if staticcheck is installed)
staticcheck ./...
```

## Architecture

### Package Structure

The project follows Go's standard project layout:

- **`cmd/`**: Entry points for the application
  - `airportCapacityCalculator.go` - Main application entry point

- **`internal/`**: Private application code (not importable by external projects)
  - `airport/` - Core airport modeling package
    - `airport.go` - Airport struct and operations
    - `runway/` - Runway-specific modeling
      - `runway.go` - Runway struct with designation, bearing, and length properties

- **`testdata/`**: Test fixtures and data files (empty currently)

### Domain Model

The domain is organized around aviation concepts:

1. **Airport** (`internal/airport`): Represents an airport facility
   - Contains `Name` field
   - Intended to aggregate runways and calculate capacity

2. **Runway** (`internal/airport/runway`): Represents physical runways
   - `RunwayDesignation`: Runway identifier (e.g., "09L", "27R")
   - `TrueBearing`: Runway direction in degrees
   - `LengthMeters`: Physical runway length
   - `MinimumSeparation`: Per-runway minimum separation time between operations

### Design Principles

- Uses Go's internal package pattern to enforce encapsulation
- Domain-driven design with aviation domain concepts
- Currently no external dependencies beyond Go standard library

## Simulation Policy System

The simulation framework uses a **Policy** pattern to model real-world constraints and operational rules that affect airport capacity. Policies are modular, reusable components that modify simulation state.

### Policy Interface

All policies must implement the `Policy` interface defined in `internal/simulation/simulation.go`:

```go
type Policy interface {
    Name() string
    Apply(ctx context.Context, state any, logger *slog.Logger) error
}
```

### How to Add a New Policy

#### Step 1: Create Policy Implementation File

Create a new file in `internal/simulation/policy/` (e.g., `yourpolicy.go`):

```go
package policy

import (
    "context"
    "fmt"
    "log/slog"
)

// YourPolicy represents your policy's purpose
type YourPolicy struct {
    // Add fields for configuration
    someParameter string
}

// NewYourPolicy creates a new instance with validation
func NewYourPolicy(param string) (*YourPolicy, error) {
    // Add validation logic
    if param == "" {
        return nil, fmt.Errorf("parameter cannot be empty")
    }

    return &YourPolicy{
        someParameter: param,
    }, nil
}

// Name returns the policy name for logging
func (p *YourPolicy) Name() string {
    return "YourPolicy"
}

// Apply modifies the simulation state
func (p *YourPolicy) Apply(ctx context.Context, state any, logger *slog.Logger) error {
    // Type assert to get the state methods you need
    simState, ok := state.(interface {
        GetOperatingHours() float32
        SetOperatingHours(float32)
        // Add other getters/setters as needed
    })

    if !ok {
        return fmt.Errorf("invalid state type for YourPolicy")
    }

    logger.DebugContext(ctx, "Applying your policy", "param", p.someParameter)

    // Implement your policy logic here
    currentHours := simState.GetOperatingHours()
    if currentHours == 0 {
        currentHours = HoursPerYear // Default to full year
    }

    // Modify state based on your policy
    adjustedHours := currentHours * 0.9 // Example: 10% reduction
    simState.SetOperatingHours(adjustedHours)

    logger.InfoContext(ctx, "Your policy applied",
        "hours_before", currentHours,
        "hours_after", adjustedHours)

    return nil
}
```

#### Step 2: Add Tests

Create `yourpolicy_test.go` in the same directory:

```go
package policy

import (
    "context"
    "log/slog"
    "testing"
)

func TestYourPolicy_Apply(t *testing.T) {
    // Test implementation
    policy, err := NewYourPolicy("test")
    if err != nil {
        t.Fatalf("Failed to create policy: %v", err)
    }

    // Add test cases
}
```

#### Step 3: Add Convenience Method to Simulation

In `internal/simulation/simulation.go`, add a method to easily attach your policy:

```go
// AddYourPolicy adds your policy to the simulation
func (s *Simulation) AddYourPolicy(param string) (*Simulation, error) {
    p, err := policy.NewYourPolicy(param)
    if err != nil {
        return nil, err
    }
    return s.AddPolicy(p), nil
}
```

#### Step 4: Export Types/Constants (if needed)

If your policy uses custom types or constants that users need, export them in `simulation.go`:

```go
// Type aliases for convenience
type (
    MaintenanceSchedule = policy.MaintenanceSchedule
    RotationStrategy    = policy.RotationStrategy
    YourPolicyConfig    = policy.YourPolicyConfig // Add your type
)

// Your policy constants (if any)
const (
    YourConstant = policy.YourConstant
)
```

#### Step 5: Use in Application

In `cmd/airportCapacityCalculator.go` or your simulation code:

```go
sim, err := simulation.NewSimulation(exampleAirport, logger).
    AddYourPolicy("parameter")
if err != nil {
    panic(err)
}
capacity, err := sim.Run(context.Background())
```

### Existing Policies

**CurfewPolicy** (`internal/simulation/policy/curfew.go`):
- Restricts operations during specified time ranges
- Reduces annual operating hours proportionally
- Usage: `.AddCurfewPolicy(startTime, endTime)`

**MaintenancePolicy** (`internal/simulation/policy/maintenance.go`):
- Schedules recurring runway maintenance
- Reduces operating hours based on frequency and duration
- Usage: `.AddMaintenancePolicy(MaintenanceSchedule{...})`

**RunwayRotationPolicy** (`internal/simulation/policy/rotation.go`):
- Implements rotation strategies (NoRotation, TimeBasedRotation, etc.)
- Applies efficiency multipliers based on strategy
- Usage: `.RunwayRotationPolicy(simulation.TimeBasedRotation)`

### SimulationState Interface

Policies interact with `SimulationState` which has these methods:

- `GetOperatingHours() float32` / `SetOperatingHours(float32)` - Total operating hours
- `GetAvailableRunways() []airport.Runway` / `SetAvailableRunways([]airport.Runway)` - Available runways

You can extend `SimulationState` in `internal/simulation/simulation.go` if your policy needs additional state.
