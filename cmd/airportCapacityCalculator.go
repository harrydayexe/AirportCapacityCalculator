package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/airport"
	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation"
	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/policy"
)

func main() {
	// Parse command-line flags
	verbose := flag.Bool("verbose", false, "Enable verbose logging (info level and above)")
	flag.Parse()

	// Configure logger based on verbose flag
	logLevel := slog.LevelError
	if *verbose {
		logLevel = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel}))

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

	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Airport Capacity Calculator - Comprehensive Demonstration   ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Airport: Metropolitan International Airport (KMIA)")
	fmt.Println("Runway Configuration:")
	fmt.Println("  • 09L/27R: 3,685m (12,090ft) - Heavy aircraft, 75s separation")
	fmt.Println("  • 09R/27L: 3,380m (11,090ft) - Standard aircraft, 60s separation")
	fmt.Println("  • 18/36:   2,743m (9,000ft)  - Regional/domestic, 50s separation (crossing)")
	fmt.Println("  • 08/26:   2,438m (8,000ft)  - Regional/GA, 45s separation")
	fmt.Println()

	// Define operational constraints
	curfewStart := time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC)
	curfewEnd := time.Date(2024, 1, 2, 6, 0, 0, 0, time.UTC)

	// Scenario 1: Full Configuration - All Policies Applied
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("Scenario 1: REALISTIC OPERATIONS - All Constraints Applied")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("Constraints:")
	fmt.Println("  • Curfew: 11 PM - 6 AM daily (7 hours)")
	fmt.Println("  • Wind: Westerly 15kt (270°) - favoring 27 operations")
	fmt.Println("  • Rotation: Preferential runway (noise abatement)")
	fmt.Println("  • Maintenance: 09R - Monthly 8hr maintenance windows")
	fmt.Println("  • Gates: 50 gates, 45min turnaround")
	fmt.Println("  • Taxi: 8min average (5min in, 3min out)")
	fmt.Println()

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

	fmt.Println("───────────────────────────────────────────────────────────────")
	fmt.Printf("RESULT: Annual Capacity: %d movements\n", int(capacity1))
	fmt.Printf("        Daily Average: %d movements\n", int(capacity1)/365)
	fmt.Printf("        Peak Hour Estimate: %d movements\n", int(capacity1)/365/17) // 17 operating hours
	fmt.Println()

	// Scenario 2: Theoretical Maximum (No Constraints)
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("Scenario 2: THEORETICAL MAXIMUM - No Constraints")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("Configuration:")
	fmt.Println("  • No curfew (24/7 operations)")
	fmt.Println("  • Calm wind (all runways forward direction)")
	fmt.Println("  • No rotation penalty")
	fmt.Println("  • No maintenance")
	fmt.Println("  • No gate constraints")
	fmt.Println("  • No taxi time overhead")
	fmt.Println()

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

	fmt.Println("───────────────────────────────────────────────────────────────")
	fmt.Printf("RESULT: Annual Capacity: %d movements\n", int(capacity2))
	fmt.Printf("        Daily Average: %d movements\n", int(capacity2)/365)
	fmt.Printf("        Peak Hour Estimate: %d movements\n", int(capacity2)/365/24)
	fmt.Println()

	// Scenario 3: Wind Impact Analysis
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("Scenario 3: WIND IMPACT ANALYSIS")
	fmt.Println("═══════════════════════════════════════════════════════════════")

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
		fmt.Printf("%s Wind: %.0fkt @ %.0f° - %s\n", scenario.name, scenario.speed, scenario.direction, scenario.desc)

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
		fmt.Printf("  → Capacity: %d movements (daily avg: %d)\n", int(capacity), int(capacity)/365)
	}
	fmt.Println()

	// Scenario 4: Maintenance Impact
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("Scenario 4: MAINTENANCE SCHEDULING COMPARISON")
	fmt.Println("═══════════════════════════════════════════════════════════════")

	// Simple maintenance
	fmt.Println("Simple Maintenance (no coordination):")
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
	fmt.Printf("  → Capacity: %d movements\n", int(capacity4a))

	// Intelligent maintenance
	fmt.Println("Intelligent Maintenance (curfew-aware):")
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
	fmt.Printf("  → Capacity: %d movements (improvement: %d)\n", int(capacity4b), int(capacity4b-capacity4a))
	fmt.Println()

	// Scenario 5: Dynamic Wind Patterns
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("Scenario 5: DYNAMIC WIND PATTERNS")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("Demonstrating time-varying wind conditions using ScheduledWindPolicy")
	fmt.Println()

	// Sub-scenario 5a: Diurnal wind pattern (daily cycle)
	fmt.Println("5a. Diurnal Wind Pattern (7-day cycle):")
	fmt.Println("    Morning calm → Afternoon westerly builds → Evening decrease")
	fmt.Println("    06:00: 5kt | 15:00: 20kt | 21:00: 10kt | 00:00: calm")

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

	fmt.Printf("  → Capacity: %d movements (daily avg: %d)\n", int(capacity5a), int(capacity5a)/365)
	fmt.Println()

	// Sub-scenario 5b: Frontal passage (abrupt wind shift)
	fmt.Println("5b. Frontal Passage (cold front):")
	fmt.Println("    Pre-frontal: Southerly 10kt → Post-frontal: Westerly 25kt")

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

	fmt.Printf("  → Capacity: %d movements (daily avg: %d)\n", int(capacity5b), int(capacity5b)/365)
	fmt.Println()

	// Sub-scenario 5c: Seasonal wind variation
	fmt.Println("5c. Seasonal Wind Pattern:")
	fmt.Println("    Winter: 15kt/270° | Spring: 10kt/180° | Summer: 5kt/90° | Fall: 12kt/225°")

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

	fmt.Printf("  → Capacity: %d movements (daily avg: %d)\n", int(capacity5c), int(capacity5c)/365)
	fmt.Println()

	// Sub-scenario 5d: Linear wind transition
	fmt.Println("5d. Linear Wind Transition (gradual change):")
	fmt.Println("    10kt/90° → 30kt/180° over 4 hours (5 steps)")

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

	fmt.Printf("  → Capacity: %d movements (daily avg: %d)\n", int(capacity5d), int(capacity5d)/365)
	fmt.Println()

	fmt.Println("Comparison:")
	fmt.Printf("  Static Westerly 15kt: %d movements\n", int(windResults[1]))
	fmt.Printf("  Diurnal Pattern (avg 15kt): %d movements\n", int(capacity5a))
	diffPercent := int((float32(windResults[1])-capacity5a)/float32(windResults[1])*100)
	if capacity5a > windResults[1] {
		diffPercent = int((capacity5a-float32(windResults[1]))/capacity5a*100)
	}
	fmt.Printf("  Difference: %d%%\n", diffPercent)
	fmt.Println()

	// Summary
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("CAPACITY SUMMARY")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Printf("Theoretical Maximum (24/7, optimal): %d movements\n", int(capacity2))
	fmt.Printf("Realistic Operations (all constraints): %d movements\n", int(capacity1))
	fmt.Printf("Capacity Utilization: %d%%\n", int(float32(capacity1)/float32(capacity2)*100))
	fmt.Println()
	fmt.Println("Primary Limiting Factors:")
	capacityLoss := capacity2 - capacity1
	fmt.Printf("  Total capacity loss: %d movements (%d%%)\n", int(capacityLoss), int(capacityLoss/capacity2*100))
	fmt.Println("  • Curfew (7hrs daily): ~29% time reduction")
	fmt.Println("  • Rotation policy: ~10% efficiency reduction")
	fmt.Println("  • Gate/taxi constraints: Variable based on demand")
	fmt.Println("  • Maintenance: ~1-2% when scheduled intelligently")
	fmt.Println("  • Wind: 0-15% depending on conditions")
	fmt.Println()
	fmt.Println("Wind Impact Range:")
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
	fmt.Printf("  Best wind conditions: %d movements\n", int(maxWind))
	fmt.Printf("  Worst wind conditions: %d movements\n", int(minWind))
	fmt.Printf("  Range: %d movements (%d%%)\n", int(maxWind-minWind), int((maxWind-minWind)/maxWind*100))
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("Simulation complete!")
	fmt.Println("═══════════════════════════════════════════════════════════════")
}
