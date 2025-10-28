// Package runway provides runway modeling and calculations.
package runway

// Runway represents a physical runway with all operational parameters.
type Runway struct {
	RunwayDesignation string  // Runway designation (e.g., "09L", "27R")
	TrueBearing        float64 // True bearing of the runway in degrees
	LengthMeters       float64 // Length of the runway in meters
}
