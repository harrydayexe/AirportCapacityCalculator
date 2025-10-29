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
	airport := airport.Airport{
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
		},
		MinimumSeparation: 60 * time.Second,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	sim := simulation.NewSimulation(airport, logger).AddCurfewPolicy(time.Now(), time.Now().Add(time.Hour*4))

	logger.Info("Basic Simulation. Single runway, 60 second separation")
	capacity, err := sim.Run(context.Background())
	if err != nil {
		panic(err)
	}

	logger.Info("Calculated annual capacity (movements per year):", "capacity", capacity)
}
