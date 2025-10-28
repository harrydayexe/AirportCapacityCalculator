# Airport Capacity Calculator

A niche data experiment to calculate the theoretical maximum capacity of an airport's movements.

## Getting Started

### Prerequisites

- Go 1.24.4 or higher

### Building and Running

```bash
# Build the main binary
go build -o airportCapacityCalculator ./cmd/airportCapacityCalculator.go

# Run the application
go run ./cmd/airportCapacityCalculator.go
```

## Running Simulations

### Pre-configured Simulations

The project currently includes the following simulation:

- **Basic Simulation** (`cmd/airportCapacityCalculator.go`): A simple single-runway simulation in a "perfect world" scenario with 60-second separation between aircraft movements.

To view the pre-configured simulation code:

```bash
# View the main simulation
cat cmd/airportCapacityCalculator.go
```

### Running a Simulation

To run the default Basic Simulation:

```bash
go run ./cmd/airportCapacityCalculator.go
```

This will:
1. Create an example airport with one runway
2. Set a minimum separation of 60 seconds between movements
3. Calculate the theoretical annual capacity (movements per year)
4. Output the results to the console

Example output:
```
Basic Simulation. Single runway, 60 second separation
time=... level=INFO msg="Starting Basic Simulation"
time=... level=INFO msg="Single runway available, calculating capacity"
...
Calculated annual capacity (movements per year): 525600
```

## Project Structure

```
.
├── cmd/
│   └── airportCapacityCalculator.go    # Main application entry point
├── internal/
│   ├── airport/
│   │   └── airport.go                  # Airport model and operations
│   └── simulation/
│       ├── simulation.go               # Simulation interface
│       └── basic-sim.go                # Basic simulation implementation
├── testdata/                           # Test fixtures and data files
├── CLAUDE.md                           # Claude Code project instructions
└── README.md                           # This file
```

### Package Overview

- **`cmd/`**: Application entry points
- **`internal/airport/`**: Airport domain models (Airport, Runway)
- **`internal/simulation/`**: Simulation implementations
  - `Simulation` interface: Defines the contract for all simulations
  - `BasicSim`: Perfect world scenario with no downtime or inefficiencies

## Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to contribute to this project.

## License

This project is currently unlicensed.