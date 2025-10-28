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
   - `RunwayDesigntation`: Runway identifier (e.g., "09L", "27R")
   - `TrueBearing`: Runway direction in degrees
   - `LengthMeters`: Physical runway length

Note: The runway package currently has a typo in the field name "RunwayDesigntation" that should be "RunwayDesignation".

### Design Principles

- Uses Go's internal package pattern to enforce encapsulation
- Domain-driven design with aviation domain concepts
- Currently no external dependencies beyond Go standard library
