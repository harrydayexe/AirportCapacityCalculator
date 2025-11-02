// Package airport provides combined airport modeling and calculations.
package airport

// Airport represents a physical airport with all its subcomponents.
type Airport struct {
	Name                string                // The commercial name of the airport
	IATACode            string                // The IATA code of the Airport
	ICAOCode            string                // The ICAO code of the Airport
	City                string                // The city where the airport is located
	Country             string                // The country where the airport is located
	Runways             []Runway              // A list of runways at the Airport
	RunwayCompatibility *RunwayCompatibility  // Optional compatibility graph defining which runways can operate simultaneously (nil means all runways compatible)
}
