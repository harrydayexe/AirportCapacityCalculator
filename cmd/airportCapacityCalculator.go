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
	// Create an airport with multiple runways to demonstrate rotation policies
	exampleAirport := airport.Airport{
		Name:     "Example Airport",
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
			},
			{
				RunwayDesignation: "09R",
				TrueBearing:       90.0,
				LengthMeters:      3000.0,
				WidthMeters:       45.0,
				SurfaceType:       airport.Asphalt,
			},
			{
				RunwayDesignation: "18",
				TrueBearing:       180.0,
				LengthMeters:      2500.0,
				WidthMeters:       45.0,
				SurfaceType:       airport.Asphalt,
			},
		},
		MinimumSeparation: 60 * time.Second,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	// Define a curfew from 11 PM to 6 AM (7 hours)
	curfewStart := time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC)
	curfewEnd := time.Date(2024, 1, 2, 6, 0, 0, 0, time.UTC)

	logger.Info("=== Runway Rotation Policy Demonstration ===")
	logger.Info("Airport Configuration", "runways", len(exampleAirport.Runways), "separation", "60s", "curfew", "7h daily")
	logger.Info("")

	// Scenario 1: No rotation (baseline)
	logger.Info("Scenario 1: No Rotation (baseline - maximum efficiency)")
	sim1 := simulation.NewSimulation(exampleAirport, logger).
		AddCurfewPolicy(curfewStart, curfewEnd).
		RunwayRotationPolicy(simulation.NoRotation)
	capacity1, err := sim1.Run(context.Background())
	if err != nil {
		panic(err)
	}
	logger.Info("Result", "capacity", int(capacity1), "strategy", "NoRotation")
	logger.Info("")

	// Scenario 2: Time-based rotation
	logger.Info("Scenario 2: Time-Based Rotation (5% capacity reduction)")
	sim2 := simulation.NewSimulation(exampleAirport, logger).
		AddCurfewPolicy(curfewStart, curfewEnd).
		RunwayRotationPolicy(simulation.TimeBasedRotation)
	capacity2, err := sim2.Run(context.Background())
	if err != nil {
		panic(err)
	}
	logger.Info("Result", "capacity", int(capacity2), "strategy", "TimeBasedRotation")
	logger.Info("")

	// Scenario 3: Balanced rotation
	logger.Info("Scenario 3: Balanced Rotation (10% capacity reduction)")
	sim3 := simulation.NewSimulation(exampleAirport, logger).
		AddCurfewPolicy(curfewStart, curfewEnd).
		RunwayRotationPolicy(simulation.BalancedRotation)
	capacity3, err := sim3.Run(context.Background())
	if err != nil {
		panic(err)
	}
	logger.Info("Result", "capacity", int(capacity3), "strategy", "BalancedRotation")
	logger.Info("")

	// Scenario 4: Noise-optimized rotation
	logger.Info("Scenario 4: Noise-Optimized Rotation (20% capacity reduction)")
	sim4 := simulation.NewSimulation(exampleAirport, logger).
		AddCurfewPolicy(curfewStart, curfewEnd).
		RunwayRotationPolicy(simulation.NoiseOptimizedRotation)
	capacity4, err := sim4.Run(context.Background())
	if err != nil {
		panic(err)
	}
	logger.Info("Result", "capacity", int(capacity4), "strategy", "NoiseOptimizedRotation")
	logger.Info("")

	println("=== Capacity Comparison ===")
	println("NoRotation Capacity:", int(capacity1))
	println("TimeBasedRotation Capacity:", int(capacity2))
	println("BalancedRotation Capacity:", int(capacity3))
	println("NoiseOptimizedRotation Capacity:", int(capacity4))
	println("")
	println("=== Impact Analysis ===")
	println("TimeBasedRotation Reduction:", int(capacity1-capacity2))
	println("BalancedRotation Reduction:", int(capacity1-capacity3))
	println("NoiseOptimizedRotation Reduction:", int(capacity1-capacity4))
}
