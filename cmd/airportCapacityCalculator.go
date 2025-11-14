package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/airport"
	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation"
	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/policy"
)

func main() {
	// Create a realistic major international airport configuration
	// Inspired by airports like LAX, with parallel runways and a crossing runway
	majorAirport := airport.Airport{
		Name:     "Metropolitan International Airport",
		IATACode: "MIA",
		ICAOCode: "KMIA",
		City:     "Metropolitan City",
		Country:  "United States",
		Runways: []airport.Runway{
			// North parallel runway complex (09L/27R)
			{
				RunwayDesignation:   "09L",
				TrueBearing:         86.0, // Slightly off from magnetic east
				LengthMeters:        3685.0, // 12,090 ft - typical for wide-body aircraft
				WidthMeters:         60.0,
				SurfaceType:         airport.Asphalt,
				ElevationMeters:     15.0,
				GradientPercent:     0.1,
				CrosswindLimitKnots: 38.0, // Wide-body aircraft on dry runway
				TailwindLimitKnots:  10.0,
				MinimumSeparation:   75 * time.Second, // Heavy aircraft separation
			},
			// South parallel runway (09R/27L)
			{
				RunwayDesignation:   "09R",
				TrueBearing:         86.0,
				LengthMeters:        3380.0, // 11,090 ft
				WidthMeters:         45.0,
				SurfaceType:         airport.Asphalt,
				ElevationMeters:     12.0,
				GradientPercent:     0.05,
				CrosswindLimitKnots: 35.0, // Narrow-body typical
				TailwindLimitKnots:  10.0,
				MinimumSeparation:   60 * time.Second,
			},
			// Crossing runway for wind coverage (18/36)
			{
				RunwayDesignation:   "18",
				TrueBearing:         176.0,
				LengthMeters:        2743.0, // 9,000 ft - regional/domestic
				WidthMeters:         45.0,
				SurfaceType:         airport.Concrete,
				ElevationMeters:     14.0,
				GradientPercent:     0.15,
				CrosswindLimitKnots: 33.0,
				TailwindLimitKnots:  8.0, // Shorter runway, more conservative
				MinimumSeparation:   50 * time.Second, // Smaller aircraft
			},
			// Additional parallel (for high capacity operations)
			{
				RunwayDesignation:   "08",
				TrueBearing:         80.0,
				LengthMeters:        2438.0, // 8,000 ft - general aviation/regional
				WidthMeters:         45.0,
				SurfaceType:         airport.Asphalt,
				ElevationMeters:     10.0,
				GradientPercent:     0.08,
				CrosswindLimitKnots: 30.0,
				TailwindLimitKnots:  8.0,
				MinimumSeparation:   45 * time.Second,
			},
		},
		// Define runway compatibility - parallel runways can operate simultaneously
		// Crossing runway (18) conflicts with all east-west runways
		RunwayCompatibility: airport.NewRunwayCompatibility(map[string][]string{
			"09L": {"09R", "08"}, // Parallel runways compatible
			"09R": {"09L", "08"},
			"08":  {"09L", "09R"},
			"18":  {}, // Crossing runway - incompatible with all
		}),
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	logger.Info("╔═══════════════════════════════════════════════════════════════╗")
	logger.Info("║   Airport Capacity Calculator - Comprehensive Demonstration   ║")
	logger.Info("╚═══════════════════════════════════════════════════════════════╝")
	logger.Info("")
	logger.Info("Airport: Metropolitan International Airport (KMIA)")
	logger.Info("Runway Configuration:")
	logger.Info("  • 09L/27R: 3,685m (12,090ft) - Heavy aircraft, 75s separation")
	logger.Info("  • 09R/27L: 3,380m (11,090ft) - Standard aircraft, 60s separation")
	logger.Info("  • 18/36:   2,743m (9,000ft)  - Regional/domestic, 50s separation (crossing)")
	logger.Info("  • 08/26:   2,438m (8,000ft)  - Regional/GA, 45s separation")
	logger.Info("")

	// Define operational constraints
	curfewStart := time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC)
	curfewEnd := time.Date(2024, 1, 2, 6, 0, 0, 0, time.UTC)

	// Scenario 1: Full Configuration - All Policies Applied
	logger.Info("═══════════════════════════════════════════════════════════════")
	logger.Info("Scenario 1: REALISTIC OPERATIONS - All Constraints Applied")
	logger.Info("═══════════════════════════════════════════════════════════════")
	logger.Info("Constraints:")
	logger.Info("  • Curfew: 11 PM - 6 AM daily (7 hours)")
	logger.Info("  • Wind: Westerly 15kt (270°) - favoring 27 operations")
	logger.Info("  • Rotation: Preferential runway (noise abatement)")
	logger.Info("  • Maintenance: 09R - Monthly 8hr maintenance windows")
	logger.Info("  • Gates: 50 gates, 45min turnaround")
	logger.Info("  • Taxi: 8min average (5min in, 3min out)")
	logger.Info("")

	sim1Temp, err := simulation.NewSimulation(majorAirport, logger).
		AddCurfewPolicy(curfewStart, curfewEnd)
	if err != nil {
		panic(err)
	}

	sim1Temp, err = sim1Temp.AddWindPolicy(15, 270) // Westerly wind
	if err != nil {
		panic(err)
	}

	sim1Temp = sim1Temp.RunwayRotationPolicy(simulation.PreferentialRunway)

	// Add maintenance for 09R
	sim1Temp = sim1Temp.AddMaintenancePolicy(simulation.MaintenanceSchedule{
		RunwayDesignations: []string{"09R"},
		Duration:           8 * time.Hour,
		Frequency:          30 * 24 * time.Hour, // Monthly
	})

	// Add gate capacity constraint
	sim1Temp, err = sim1Temp.AddGateCapacityPolicy(simulation.GateCapacityConstraint{
		TotalGates:            50,
		AverageTurnaroundTime: 45 * time.Minute,
	})
	if err != nil {
		panic(err)
	}

	// Add taxi time
	sim1Temp, err = sim1Temp.AddTaxiTimePolicy(simulation.TaxiTimeConfiguration{
		AverageTaxiInTime: 5 * time.Minute,
		AverageTaxiOutTime: 3 * time.Minute,
	})
	if err != nil {
		panic(err)
	}

	capacity1, err := sim1Temp.Run(context.Background())
	if err != nil {
		panic(err)
	}

	logger.Info("───────────────────────────────────────────────────────────────")
	logger.Info("RESULT: Annual Capacity", "movements", int(capacity1))
	logger.Info("        Daily Average", "movements", int(capacity1)/365)
	logger.Info("        Peak Hour Estimate", "movements", int(capacity1)/365/17) // 17 operating hours
	logger.Info("")

	// Scenario 2: Theoretical Maximum (No Constraints)
	logger.Info("═══════════════════════════════════════════════════════════════")
	logger.Info("Scenario 2: THEORETICAL MAXIMUM - No Constraints")
	logger.Info("═══════════════════════════════════════════════════════════════")
	logger.Info("Configuration:")
	logger.Info("  • No curfew (24/7 operations)")
	logger.Info("  • Calm wind (all runways forward direction)")
	logger.Info("  • No rotation penalty")
	logger.Info("  • No maintenance")
	logger.Info("  • No gate constraints")
	logger.Info("  • No taxi time overhead")
	logger.Info("")

	sim2Temp, err := simulation.NewSimulation(majorAirport, logger).
		AddWindPolicy(0, 0) // Calm wind
	if err != nil {
		panic(err)
	}

	sim2Temp = sim2Temp.RunwayRotationPolicy(simulation.NoRotation)

	capacity2, err := sim2Temp.Run(context.Background())
	if err != nil {
		panic(err)
	}

	logger.Info("───────────────────────────────────────────────────────────────")
	logger.Info("RESULT: Annual Capacity", "movements", int(capacity2))
	logger.Info("        Daily Average", "movements", int(capacity2)/365)
	logger.Info("        Peak Hour Estimate", "movements", int(capacity2)/365/24)
	logger.Info("")

	// Scenario 3: Wind Impact Analysis
	logger.Info("═══════════════════════════════════════════════════════════════")
	logger.Info("Scenario 3: WIND IMPACT ANALYSIS")
	logger.Info("═══════════════════════════════════════════════════════════════")

	windScenarios := []struct {
		name      string
		speed     float64
		direction float64
		desc      string
	}{
		{"Calm", 0, 0, "All runways available, forward direction"},
		{"Light Westerly", 10, 270, "Favors 27 operations (reverse direction)"},
		{"Moderate Westerly", 20, 270, "Strong preference for 27 operations"},
		{"Strong Northerly", 35, 360, "Strong crosswind on east-west runways"},
		{"Southerly", 15, 180, "Favors 36 operation (crossing runway)"},
	}

	windResults := make([]float32, len(windScenarios))

	for i, scenario := range windScenarios {
		logger.Info(scenario.name+" Wind", "speed", scenario.speed, "direction", scenario.direction, "desc", scenario.desc)

		simTemp, err := simulation.NewSimulation(majorAirport, logger).
			AddCurfewPolicy(curfewStart, curfewEnd)
		if err != nil {
			panic(err)
		}

		simTemp, err = simTemp.AddWindPolicy(scenario.speed, scenario.direction)
		if err != nil {
			panic(err)
		}

		capacity, err := simTemp.Run(context.Background())
		if err != nil {
			panic(err)
		}

		windResults[i] = capacity
		logger.Info("  → Capacity", "movements", int(capacity), "daily_avg", int(capacity)/365)
	}
	logger.Info("")

	// Scenario 4: Maintenance Impact
	logger.Info("═══════════════════════════════════════════════════════════════")
	logger.Info("Scenario 4: MAINTENANCE SCHEDULING COMPARISON")
	logger.Info("═══════════════════════════════════════════════════════════════")

	// Simple maintenance
	logger.Info("Simple Maintenance (no coordination):")
	sim4aTemp, err := simulation.NewSimulation(majorAirport, logger).
		AddCurfewPolicy(curfewStart, curfewEnd)
	if err != nil {
		panic(err)
	}

	sim4aTemp, err = sim4aTemp.AddWindPolicy(15, 270)
	if err != nil {
		panic(err)
	}

	sim4aTemp = sim4aTemp.AddMaintenancePolicy(simulation.MaintenanceSchedule{
		RunwayDesignations: []string{"09L"},
		Duration:           12 * time.Hour,
		Frequency:          30 * 24 * time.Hour,
	})

	capacity4a, err := sim4aTemp.Run(context.Background())
	if err != nil {
		panic(err)
	}
	logger.Info("  → Capacity", "movements", int(capacity4a))

	// Intelligent maintenance
	logger.Info("Intelligent Maintenance (curfew-aware):")
	sim4bTemp, err := simulation.NewSimulation(majorAirport, logger).
		AddCurfewPolicy(curfewStart, curfewEnd)
	if err != nil {
		panic(err)
	}

	sim4bTemp, err = sim4bTemp.AddWindPolicy(15, 270)
	if err != nil {
		panic(err)
	}

	sim4bTemp, err = sim4bTemp.AddIntelligentMaintenancePolicy(simulation.IntelligentMaintenanceSchedule{
		RunwayDesignations:        []string{"09L"},
		Duration:                  12 * time.Hour,
		Frequency:                 30 * 24 * time.Hour,
		MinimumOperationalRunways: 2,
	})
	if err != nil {
		panic(err)
	}

	capacity4b, err := sim4bTemp.Run(context.Background())
	if err != nil {
		panic(err)
	}
	logger.Info("  → Capacity", "movements", int(capacity4b), "improvement", int(capacity4b-capacity4a))
	logger.Info("")

	// Scenario 5: Dynamic Wind Patterns
	logger.Info("═══════════════════════════════════════════════════════════════")
	logger.Info("Scenario 5: DYNAMIC WIND PATTERNS")
	logger.Info("═══════════════════════════════════════════════════════════════")
	logger.Info("Demonstrating time-varying wind conditions using ScheduledWindPolicy")
	logger.Info("")

	// Sub-scenario 5a: Diurnal wind pattern (daily cycle)
	logger.Info("5a. Diurnal Wind Pattern (7-day cycle):")
	logger.Info("    Morning calm → Afternoon westerly builds → Evening decrease")
	logger.Info("    06:00: 5kt | 15:00: 20kt | 21:00: 10kt | 00:00: calm")

	diurnalSchedule := policy.DiurnalWindPattern(
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		7,    // 7 days
		5.0,  // morning: 5kt
		20.0, // afternoon peak: 20kt
		10.0, // evening: 10kt
		270,  // westerly direction
	)

	sim5aTemp, err := simulation.NewSimulation(majorAirport, logger).
		AddCurfewPolicy(curfewStart, curfewEnd)
	if err != nil {
		panic(err)
	}

	sim5aTemp, err = sim5aTemp.AddScheduledWindPolicy(diurnalSchedule)
	if err != nil {
		panic(err)
	}

	capacity5a, err := sim5aTemp.Run(context.Background())
	if err != nil {
		panic(err)
	}

	logger.Info("  → Capacity", "movements", int(capacity5a), "daily_avg", int(capacity5a)/365)
	logger.Info("")

	// Sub-scenario 5b: Frontal passage (abrupt wind shift)
	logger.Info("5b. Frontal Passage (cold front):")
	logger.Info("    Pre-frontal: Southerly 10kt → Post-frontal: Westerly 25kt")

	frontPassageTime := time.Date(2024, 3, 15, 18, 0, 0, 0, time.UTC)
	frontalSchedule := policy.FrontalPassagePattern(
		frontPassageTime,
		10,  // pre-frontal speed (kt)
		180, // pre-frontal direction (south)
		25,  // post-frontal speed (kt)
		270, // post-frontal direction (west)
	)

	sim5bTemp, err := simulation.NewSimulation(majorAirport, logger).
		AddCurfewPolicy(curfewStart, curfewEnd)
	if err != nil {
		panic(err)
	}

	sim5bTemp, err = sim5bTemp.AddScheduledWindPolicy(frontalSchedule)
	if err != nil {
		panic(err)
	}

	capacity5b, err := sim5bTemp.Run(context.Background())
	if err != nil {
		panic(err)
	}

	logger.Info("  → Capacity", "movements", int(capacity5b), "daily_avg", int(capacity5b)/365)
	logger.Info("")

	// Sub-scenario 5c: Seasonal wind variation
	logger.Info("5c. Seasonal Wind Pattern:")
	logger.Info("    Winter: 15kt/270° | Spring: 10kt/180° | Summer: 5kt/90° | Fall: 12kt/225°")

	seasonalSchedule := policy.SeasonalWindPattern(
		2024,
		time.UTC,
		15, 10, 5, 12,   // speeds (winter, spring, summer, fall)
		270, 180, 90, 225, // directions
	)

	sim5cTemp, err := simulation.NewSimulation(majorAirport, logger).
		AddCurfewPolicy(curfewStart, curfewEnd)
	if err != nil {
		panic(err)
	}

	sim5cTemp, err = sim5cTemp.AddScheduledWindPolicy(seasonalSchedule)
	if err != nil {
		panic(err)
	}

	capacity5c, err := sim5cTemp.Run(context.Background())
	if err != nil {
		panic(err)
	}

	logger.Info("  → Capacity", "movements", int(capacity5c), "daily_avg", int(capacity5c)/365)
	logger.Info("")

	// Sub-scenario 5d: Linear wind transition
	logger.Info("5d. Linear Wind Transition (gradual change):")
	logger.Info("    10kt/90° → 30kt/180° over 4 hours (5 steps)")

	transitionStart := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	transitionSchedule, err := policy.LinearWindTransition(
		transitionStart,
		4*time.Hour, // duration
		5,           // steps
		10, 90,      // initial: 10kt from east
		30, 180,     // final: 30kt from south
	)
	if err != nil {
		panic(err)
	}

	sim5dTemp, err := simulation.NewSimulation(majorAirport, logger).
		AddCurfewPolicy(curfewStart, curfewEnd)
	if err != nil {
		panic(err)
	}

	sim5dTemp, err = sim5dTemp.AddScheduledWindPolicy(transitionSchedule)
	if err != nil {
		panic(err)
	}

	capacity5d, err := sim5dTemp.Run(context.Background())
	if err != nil {
		panic(err)
	}

	logger.Info("  → Capacity", "movements", int(capacity5d), "daily_avg", int(capacity5d)/365)
	logger.Info("")

	logger.Info("Comparison:")
	logger.Info("  Static Westerly 15kt", "movements", int(windResults[1]))
	logger.Info("  Diurnal Pattern (avg 15kt)", "movements", int(capacity5a))
	diffPercent := int((float32(windResults[1])-capacity5a)/float32(windResults[1])*100)
	if capacity5a > windResults[1] {
		diffPercent = int((capacity5a-float32(windResults[1]))/capacity5a*100)
	}
	logger.Info("  Difference", "percent", diffPercent)
	logger.Info("")

	// Summary
	logger.Info("═══════════════════════════════════════════════════════════════")
	logger.Info("CAPACITY SUMMARY")
	logger.Info("═══════════════════════════════════════════════════════════════")
	logger.Info("Theoretical Maximum (24/7, optimal)", "movements", int(capacity2))
	logger.Info("Realistic Operations (all constraints)", "movements", int(capacity1))
	logger.Info("Capacity Utilization", "percent", int(float32(capacity1)/float32(capacity2)*100))
	logger.Info("")
	logger.Info("Primary Limiting Factors:")
	capacityLoss := capacity2 - capacity1
	logger.Info("  Total capacity loss", "movements", int(capacityLoss), "percent", int(capacityLoss/capacity2*100))
	logger.Info("  • Curfew (7hrs daily): ~29% time reduction")
	logger.Info("  • Rotation policy: ~10% efficiency reduction")
	logger.Info("  • Gate/taxi constraints: Variable based on demand")
	logger.Info("  • Maintenance: ~1-2% when scheduled intelligently")
	logger.Info("  • Wind: 0-15% depending on conditions")
	logger.Info("")
	logger.Info("Wind Impact Range:")
	maxWind := windResults[0]
	minWind := windResults[0]
	for _, result := range windResults {
		if result > maxWind {
			maxWind = result
		}
		if result < minWind {
			minWind = result
		}
	}
	logger.Info("  Best wind conditions", "movements", int(maxWind))
	logger.Info("  Worst wind conditions", "movements", int(minWind))
	logger.Info("  Range", "movements", int(maxWind-minWind), "percent", int((maxWind-minWind)/maxWind*100))
	logger.Info("")
	logger.Info("═══════════════════════════════════════════════════════════════")
	logger.Info("Simulation complete! 🎉")
	logger.Info("═══════════════════════════════════════════════════════════════")
}
