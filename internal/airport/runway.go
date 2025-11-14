package airport

import "time"

// SurfaceType represents the type of surface of the runway.
type SurfaceType int

const (
	Asphalt SurfaceType = iota
	Concrete
	Grass
	Dirt
)

// Runway represents a physical runway with all operational parameters.
type Runway struct {
	RunwayDesignation  string        // Runway designation (e.g., "09L", "27R")
	TrueBearing        float64       // True bearing of the runway in degrees
	LengthMeters       float64       // Length of the runway in meters
	WidthMeters        float64       // Width of the runway in WidthMeters
	SurfaceType        SurfaceType   // Surface type of the runway (e.g., "Asphalt", "Concrete", "Grass")
	ElevationMeters    float64       // Elevation of the runway above sea level in meters
	GradientPercent    float64       // Gradient of the runway in percent
	CrosswindLimitKnots float64       // Maximum crosswind component in knots (0 = no limit)
	TailwindLimitKnots  float64       // Maximum tailwind component in knots (0 = no limit)
	MinimumSeparation  time.Duration // Minimum separation time between incoming flights
}
