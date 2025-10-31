# Contributing to Airport Capacity Calculator

Thank you for your interest in contributing to the Airport Capacity Calculator project! This guide will help you understand how to contribute effectively.

## Table of Contents
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Adding New Simulation Policies](#adding-new-simulation-policies)
- [Code Style and Standards](#code-style-and-standards)
- [Testing Guidelines](#testing-guidelines)
- [Submitting Changes](#submitting-changes)

## Getting Started

### Prerequisites
- Go 1.24.4 or later
- Basic understanding of Go modules and packages
- Familiarity with airport operations (helpful but not required)

### Setting Up Your Development Environment

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/AirportCapacityCalculator.git
   cd AirportCapacityCalculator
   ```
3. Build and test:
   ```bash
   go build ./...
   go test ./...
   ```

## Development Workflow

1. Create a feature branch from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```
2. Make your changes
3. Run tests and formatting:
   ```bash
   go test ./...
   go fmt ./...
   go vet ./...
   ```
4. Commit your changes with clear, descriptive messages
5. Push to your fork and create a pull request

## Adding New Simulation Policies

Policies are runtime components that affect simulation behavior during execution. They modify the simulation state to model real-world constraints like curfews, maintenance windows, or operational strategies.

### Policy Interface

All policies must implement the `Policy` interface defined in `internal/simulation/simulation.go:18-22`:

```go
type Policy interface {
    Name() string
    Apply(ctx context.Context, state any) error
}
```

### Step-by-Step Guide to Adding a New Policy

#### 1. Create the Policy File

Create a new file in `internal/simulation/policy/` for your policy:

```bash
internal/simulation/policy/yourpolicy.go
```

#### 2. Define the Policy Structure

Your policy should:
- Be in the `policy` package
- Have a struct that holds its configuration
- Include a constructor function
- Implement the `Policy` interface

Example template:

```go
package policy

import (
    "context"
    "fmt"
)

// YourPolicy describes what your policy does.
// Example: "WeatherPolicy models weather-related operational constraints."
type YourPolicy struct {
    // Configuration fields
    configField string
    anotherField int
}

// NewYourPolicy creates a new instance of YourPolicy.
func NewYourPolicy(configField string, anotherField int) *YourPolicy {
    return &YourPolicy{
        configField: configField,
        anotherField: anotherField,
    }
}

// Name returns the policy name for logging and identification.
func (p *YourPolicy) Name() string {
    return "YourPolicy"
}

// Apply applies the policy to the simulation state.
// This is where your policy logic goes.
func (p *YourPolicy) Apply(ctx context.Context, state any) error {
    // Type assert to access the state interface methods you need
    simState, ok := state.(interface {
        GetOperatingHours() float32
        SetOperatingHours(float32)
        GetAvailableRunways() []airport.Runway
        SetAvailableRunways([]airport.Runway)
    })

    if !ok {
        return fmt.Errorf("invalid state type for YourPolicy")
    }

    // Implement your policy logic here
    // Modify the state based on your policy's rules

    return nil
}
```

#### 3. Access SimulationState

The `state` parameter is a `SimulationState` (defined in `internal/simulation/simulation.go:38-45`) which includes:

- `Airport` - The airport being simulated
- `CurrentTime` - Current simulation time
- `AvailableRunways` - Runways currently available
- `TotalMovements` - Total movements processed
- `OperatingHours` - Total hours of operation

Use interface type assertions to access the getters/setters you need:

```go
// For operating hours
simState.(interface {
    GetOperatingHours() float32
    SetOperatingHours(float32)
})

// For runways
simState.(interface {
    GetAvailableRunways() []airport.Runway
    SetAvailableRunways([]airport.Runway)
})
```

#### 4. Add a Convenience Method to Simulation

In `internal/simulation/simulation.go`, add a convenience method for your policy:

```go
// AddYourPolicy adds your policy with the specified configuration.
func (s *Simulation) AddYourPolicy(configField string, anotherField int) *Simulation {
    p := policy.NewYourPolicy(configField, anotherField)
    return s.AddPolicy(p)
}
```

This enables the fluent API pattern:
```go
sim.AddYourPolicy("config", 42).AddCurfewPolicy(start, end)
```

#### 5. Write Tests

Create a test file `internal/simulation/policy/yourpolicy_test.go`:

```go
package policy_test

import (
    "context"
    "testing"

    "github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/policy"
)

func TestYourPolicy_Name(t *testing.T) {
    p := policy.NewYourPolicy("test", 1)
    if p.Name() != "YourPolicy" {
        t.Errorf("expected 'YourPolicy', got '%s'", p.Name())
    }
}

func TestYourPolicy_Apply(t *testing.T) {
    p := policy.NewYourPolicy("test", 1)

    // Create a mock state or use a real SimulationState
    // Test that Apply() modifies state correctly

    // Example:
    // state := &mockState{...}
    // err := p.Apply(context.Background(), state)
    // if err != nil { ... }
    // verify state was modified correctly
}
```

#### 6. Update Documentation

If you're adding type aliases or constants to `internal/simulation/simulation.go`, document them:

```go
// Type aliases for convenience
type (
    YourPolicyConfig = policy.YourPolicyConfig
)

// Your policy constants
const (
    ConfigOption1 = policy.ConfigOption1
    ConfigOption2 = policy.ConfigOption2
)
```

### Real-World Examples

Study these existing policies for reference:

1. **CurfewPolicy** (`internal/simulation/policy/curfew.go`) - Reduces operating hours
2. **MaintenancePolicy** (`internal/simulation/policy/maintenance.go`) - Schedules downtime
3. **RunwayRotationPolicy** (`internal/simulation/policy/rotation.go`) - Uses strategy pattern with constants

### Usage Example

Once implemented, your policy can be used like this:

```go
sim := simulation.NewSimulation(airport, logger).
    AddYourPolicy("config", 42).
    AddCurfewPolicy(startTime, endTime).
    Run(ctx)
```

Or using the generic `AddPolicy` method:

```go
p := policy.NewYourPolicy("config", 42)
sim := simulation.NewSimulation(airport, logger).
    AddPolicy(p).
    Run(ctx)
```

## Code Style and Standards

### General Guidelines
- Follow standard Go conventions and idioms
- Use `gofmt` to format all code
- Run `go vet` to catch common issues
- Keep functions focused and single-purpose
- Write clear, descriptive comments for exported types and functions

### Naming Conventions
- Use descriptive names that reflect aviation domain concepts
- Avoid abbreviations unless they're standard aviation terms (IATA, ICAO, etc.)
- Policy names should end with "Policy" (e.g., `CurfewPolicy`, `MaintenancePolicy`)

### Error Handling
- Return meaningful error messages
- Use `fmt.Errorf` for error context
- Don't panic except in truly unrecoverable situations

## Testing Guidelines

### Test Coverage
- Write tests for all new functionality
- Aim for meaningful test coverage, not just high percentages
- Test both success and error cases

### Test Organization
- Place tests in `*_test.go` files
- Use table-driven tests for multiple scenarios
- Use descriptive test names: `TestPolicyName_Scenario`

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for a specific package
go test ./internal/simulation/policy/...

# Run with verbose output
go test -v ./...
```

## Submitting Changes

### Pull Request Process

1. Ensure all tests pass and code is formatted
2. Update relevant documentation
3. Write a clear PR description explaining:
   - What problem does this solve?
   - What changes were made?
   - How has it been tested?
4. Link any related issues
5. Request review from maintainers

### Commit Message Format

Use clear, descriptive commit messages:

```
type(scope): brief description

Longer explanation if needed, explaining what and why,
not how.

Fixes #123
```

Types: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`

Examples:
- `feat(policy): add weather impact policy`
- `fix(simulation): correct operating hours calculation`
- `docs(contributing): add policy development guide`
- `test(runway): add tests for rotation strategy`

### Code Review

- Be responsive to feedback
- Keep discussions professional and constructive
- Be open to suggestions and alternative approaches
- Update your PR based on review comments

## Questions or Issues?

If you have questions or run into issues:
- Check existing issues on GitHub
- Review the project documentation in `/docs`
- Open a new issue with a clear description
- Tag it appropriately (bug, question, enhancement, etc.)

Thank you for contributing to Airport Capacity Calculator!

