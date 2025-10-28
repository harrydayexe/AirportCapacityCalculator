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
	var airport = airport.Airport{
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

	var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	var sim = simulation.NewBasicSim(airport, logger)

	println("Basic Simulation. Single runway, 60 second separation")
	capacity, err := sim.Run(context.Background())
	if err != nil {
		panic(err)
	}

	println("Calculated annual capacity (movements per year):", capacity)
}
