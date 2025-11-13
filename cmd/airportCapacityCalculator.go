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
				RunwayDesignation:   "09L",
				TrueBearing:         90.0,
				LengthMeters:        3000.0,
				WidthMeters:         45.0,
				SurfaceType:         airport.Asphalt,
				CrosswindLimitKnots: 35.0, // Typical for commercial aircraft
				TailwindLimitKnots:  10.0, // Typical limit
				MinimumSeparation:   60 * time.Second,
			},
			{
				RunwayDesignation:   "09R",
				TrueBearing:         90.0,
				LengthMeters:        3000.0,
				WidthMeters:         45.0,
				SurfaceType:         airport.Asphalt,
				CrosswindLimitKnots: 35.0,
				TailwindLimitKnots:  10.0,
				MinimumSeparation:   60 * time.Second,
			},
			{
				RunwayDesignation:   "18",
				TrueBearing:         180.0,
				LengthMeters:        2500.0,
				WidthMeters:         45.0,
				SurfaceType:         airport.Asphalt,
				CrosswindLimitKnots: 35.0,
				TailwindLimitKnots:  10.0,
				MinimumSeparation:   60 * time.Second,
			},
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Define a curfew from 11 PM to 6 AM (7 hours)
	curfewStart := time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC)
	curfewEnd := time.Date(2024, 1, 2, 6, 0, 0, 0, time.UTC)

	logger.Info("=== Runway Rotation Policy Demonstration ===")
	logger.Info("Airport Configuration", "runways", len(exampleAirport.Runways), "separation", "60s", "curfew", "7h daily")
	logger.Info("")

	// Scenario 1: No rotation (baseline)
	logger.Info("Scenario 1: No Rotation (baseline - maximum efficiency)")
	sim1, err := simulation.NewSimulation(exampleAirport, logger).
		AddCurfewPolicy(curfewStart, curfewEnd)
	if err != nil {
		panic(err)
	}
	sim1 = sim1.RunwayRotationPolicy(simulation.NoRotation)
	capacity1, err := sim1.Run(context.Background())
	if err != nil {
		panic(err)
	}
	logger.Info("Result", "capacity", int(capacity1), "strategy", "NoRotation")
	logger.Info("")

	// Scenario 2: Time-based rotation
	logger.Info("Scenario 2: Time-Based Rotation (5% capacity reduction)")
	sim2, err := simulation.NewSimulation(exampleAirport, logger).
		AddCurfewPolicy(curfewStart, curfewEnd)
	if err != nil {
		panic(err)
	}
	sim2 = sim2.RunwayRotationPolicy(simulation.TimeBasedRotation)
	capacity2, err := sim2.Run(context.Background())
	if err != nil {
		panic(err)
	}
	logger.Info("Result", "capacity", int(capacity2), "strategy", "TimeBasedRotation")
	logger.Info("")

	// Scenario 3: Preferential runway
	logger.Info("Scenario 3: Preferential Runway (10% capacity reduction)")
	sim3, err := simulation.NewSimulation(exampleAirport, logger).
		AddCurfewPolicy(curfewStart, curfewEnd)
	if err != nil {
		panic(err)
	}
	sim3 = sim3.RunwayRotationPolicy(simulation.PreferentialRunway)
	capacity3, err := sim3.Run(context.Background())
	if err != nil {
		panic(err)
	}
	logger.Info("Result", "capacity", int(capacity3), "strategy", "PreferentialRunway")
	logger.Info("")

	// Scenario 4: Noise-optimized rotation
	logger.Info("Scenario 4: Noise-Optimized Rotation (20% capacity reduction)")
	sim4, err := simulation.NewSimulation(exampleAirport, logger).
		AddCurfewPolicy(curfewStart, curfewEnd)
	if err != nil {
		panic(err)
	}
	sim4 = sim4.RunwayRotationPolicy(simulation.NoiseOptimizedRotation)
	capacity4, err := sim4.Run(context.Background())
	if err != nil {
		panic(err)
	}
	logger.Info("Result", "capacity", int(capacity4), "strategy", "NoiseOptimizedRotation")
	logger.Info("")

	logger.Info("=== Capacity Comparison ===")
	logger.Info("NoRotation Capacity", "capacity", int(capacity1))
	logger.Info("TimeBasedRotation Capacity", "capacity", int(capacity2))
	logger.Info("PreferentialRunway Capacity", "capacity", int(capacity3))
	logger.Info("NoiseOptimizedRotation Capacity", "capacity", int(capacity4))
	logger.Info("")
	logger.Info("=== Impact Analysis ===")
	logger.Info("TimeBasedRotation Reduction", "reduction", int(capacity1-capacity2))
	logger.Info("PreferentialRunway Reduction", "reduction", int(capacity1-capacity3))
	logger.Info("NoiseOptimizedRotation Reduction", "reduction", int(capacity1-capacity4))
	logger.Info("")

	// Wind Policy Demonstration
	logger.Info("=== Wind Policy Demonstration ===")
	logger.Info("Demonstrating how wind affects runway usability and direction selection")
	logger.Info("")

	// Scenario 5: Westerly wind (270°) - favors runways 27 (reverse of 09)
	logger.Info("Scenario 5: Westerly wind (270° at 15kt)")
	logger.Info("Expected: Runways 09L/09R will operate in reverse direction (as 27L/27R)")
	sim5Temp, err := simulation.NewSimulation(exampleAirport, logger).
		AddCurfewPolicy(curfewStart, curfewEnd)
	if err != nil {
		panic(err)
	}
	sim5, err := sim5Temp.AddWindPolicy(15, 270) // 15kt westerly wind
	if err != nil {
		panic(err)
	}
	capacity5, err := sim5.Run(context.Background())
	if err != nil {
		panic(err)
	}
	logger.Info("Result", "capacity", int(capacity5), "wind", "270° 15kt")
	logger.Info("")

	// Scenario 6: Strong northerly crosswind (360°) - may exclude runway 18
	logger.Info("Scenario 6: Strong northerly crosswind (360° at 30kt)")
	logger.Info("Expected: Runways may operate with significant crosswind component")
	sim6Temp, err := simulation.NewSimulation(exampleAirport, logger).
		AddCurfewPolicy(curfewStart, curfewEnd)
	if err != nil {
		panic(err)
	}
	sim6, err := sim6Temp.AddWindPolicy(30, 360) // 30kt northerly wind
	if err != nil {
		panic(err)
	}
	capacity6, err := sim6.Run(context.Background())
	if err != nil {
		panic(err)
	}
	logger.Info("Result", "capacity", int(capacity6), "wind", "360° 30kt")
	logger.Info("")

	// Scenario 7: Calm wind (baseline for comparison)
	logger.Info("Scenario 7: Calm wind (no wind constraints)")
	sim7Temp, err := simulation.NewSimulation(exampleAirport, logger).
		AddCurfewPolicy(curfewStart, curfewEnd)
	if err != nil {
		panic(err)
	}
	sim7, err := sim7Temp.AddWindPolicy(0, 0) // Calm wind
	if err != nil {
		panic(err)
	}
	capacity7, err := sim7.Run(context.Background())
	if err != nil {
		panic(err)
	}
	logger.Info("Result", "capacity", int(capacity7), "wind", "calm")
	logger.Info("")

	logger.Info("=== Wind Impact Comparison ===")
	logger.Info("Calm wind capacity", "capacity", int(capacity7))
	logger.Info("Westerly wind (270° 15kt) capacity", "capacity", int(capacity5))
	logger.Info("Northerly wind (360° 30kt) capacity", "capacity", int(capacity6))
	logger.Info("")
	logger.Info("Note: Wind changes runway direction but doesn't necessarily reduce capacity")
	logger.Info("Capacity reduction occurs only if wind exceeds crosswind/tailwind limits")
}
