// THe package simulation defines the Simulation interface for running simulations.
package simulation

import "context"

// Simulation represents a simulation that can be run.
type Simulation interface {
	Run(context.Context) (float32, error) // Run executes the simulation and returns a result (max movements per year) and an error if any.
}
