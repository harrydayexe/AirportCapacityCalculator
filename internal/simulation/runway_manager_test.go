package simulation

import (
	"sync"
	"testing"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/airport"
	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/event"
)

func createTestRunways() []airport.Runway {
	return []airport.Runway{
		{
			RunwayDesignation: "09L",
			TrueBearing:       90,
			LengthMeters:      3000,
			MinimumSeparation: 90 * time.Second,
		},
		{
			RunwayDesignation: "09R",
			TrueBearing:       90,
			LengthMeters:       3200,
			MinimumSeparation: 90 * time.Second,
		},
		{
			RunwayDesignation: "18",
			TrueBearing:       180,
			LengthMeters:      2800,
			MinimumSeparation: 90 * time.Second,
		},
	}
}

func TestNewRunwayManager(t *testing.T) {
	runways := createTestRunways()
	rm := NewRunwayManager(runways, nil)

	if rm == nil {
		t.Fatal("Expected non-nil RunwayManager")
	}

	// All runways should be active initially
	config := rm.GetActiveConfiguration()
	if len(config) != len(runways) {
		t.Errorf("Expected %d active runways, got %d", len(runways), len(config))
	}

	// Verify each runway is in the configuration
	for _, runway := range runways {
		info, exists := config[runway.RunwayDesignation]
		if !exists {
			t.Errorf("Runway %s not in active configuration", runway.RunwayDesignation)
			continue
		}
		if info.OperationType != event.Mixed {
			t.Errorf("Expected Mixed operation type, got %v", info.OperationType)
		}
		if info.Direction != event.Forward {
			t.Errorf("Expected Forward direction, got %v", info.Direction)
		}
	}
}

func TestRunwayManager_OnRunwayUnavailable(t *testing.T) {
	runways := createTestRunways()
	rm := NewRunwayManager(runways, nil)

	// Mark one runway unavailable
	rm.OnRunwayUnavailable("09L")

	config := rm.GetActiveConfiguration()
	if len(config) != 2 {
		t.Errorf("Expected 2 active runways, got %d", len(config))
	}

	if _, exists := config["09L"]; exists {
		t.Error("09L should not be active after marking unavailable")
	}
}

func TestRunwayManager_OnRunwayAvailable(t *testing.T) {
	runways := createTestRunways()
	rm := NewRunwayManager(runways, nil)

	// Mark unavailable then available again
	rm.OnRunwayUnavailable("09L")
	rm.OnRunwayAvailable("09L")

	config := rm.GetActiveConfiguration()
	if len(config) != 3 {
		t.Errorf("Expected 3 active runways, got %d", len(config))
	}

	if _, exists := config["09L"]; !exists {
		t.Error("09L should be active after marking available")
	}
}

func TestRunwayManager_OnCurfewChanged(t *testing.T) {
	runways := createTestRunways()
	rm := NewRunwayManager(runways, nil)

	// Activate curfew
	rm.OnCurfewChanged(true)

	config := rm.GetActiveConfiguration()
	if len(config) != 0 {
		t.Errorf("Expected 0 active runways during curfew, got %d", len(config))
	}

	// Deactivate curfew
	rm.OnCurfewChanged(false)

	config = rm.GetActiveConfiguration()
	if len(config) != 3 {
		t.Errorf("Expected 3 active runways after curfew, got %d", len(config))
	}
}

func TestRunwayManager_ConcurrentNotifications(t *testing.T) {
	runways := createTestRunways()
	rm := NewRunwayManager(runways, nil)

	var wg sync.WaitGroup
	numGoroutines := 50

	// Concurrent availability changes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			runway := runways[id%len(runways)].RunwayDesignation

			// Alternate between available and unavailable
			if id%2 == 0 {
				rm.OnRunwayUnavailable(runway)
			} else {
				rm.OnRunwayAvailable(runway)
			}
		}(i)
	}

	wg.Wait()

	// Should not panic and should return a valid configuration
	config := rm.GetActiveConfiguration()
	if config == nil {
		t.Error("Expected non-nil configuration")
	}
}

func TestRunwayManager_ConcurrentReads(t *testing.T) {
	runways := createTestRunways()
	rm := NewRunwayManager(runways, nil)

	var wg sync.WaitGroup
	numReaders := 100

	// Start many concurrent readers
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			config := rm.GetActiveConfiguration()
			if config == nil {
				t.Error("Expected non-nil configuration")
			}
		}()
	}

	// Also do some writes
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			rm.OnRunwayUnavailable("09L")
			time.Sleep(1 * time.Millisecond)
			rm.OnRunwayAvailable("09L")
		}
	}()

	wg.Wait()
}

func TestRunwayManager_ConcurrentReadsDuringWrites(t *testing.T) {
	runways := createTestRunways()
	rm := NewRunwayManager(runways, nil)

	var wg sync.WaitGroup
	stopReaders := make(chan struct{})

	// Start continuous readers
	numReaders := 20
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopReaders:
					return
				default:
					config := rm.GetActiveConfiguration()
					// Verify config is valid
					if config == nil {
						t.Error("Got nil configuration during concurrent access")
					}
				}
			}
		}()
	}

	// Perform many writes
	numWrites := 100
	for i := 0; i < numWrites; i++ {
		runway := runways[i%len(runways)].RunwayDesignation
		rm.OnRunwayUnavailable(runway)
		rm.OnRunwayAvailable(runway)
	}

	// Stop readers
	close(stopReaders)
	wg.Wait()
}

func TestRunwayManager_ConcurrentCurfewAndAvailability(t *testing.T) {
	runways := createTestRunways()
	rm := NewRunwayManager(runways, nil)

	var wg sync.WaitGroup

	// Goroutine toggling curfew
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			rm.OnCurfewChanged(i%2 == 0)
			time.Sleep(1 * time.Millisecond)
		}
	}()

	// Goroutines toggling runway availability
	for _, runway := range runways {
		wg.Add(1)
		go func(rwy string) {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				if i%2 == 0 {
					rm.OnRunwayUnavailable(rwy)
				} else {
					rm.OnRunwayAvailable(rwy)
				}
				time.Sleep(1 * time.Millisecond)
			}
		}(runway.RunwayDesignation)
	}

	wg.Wait()

	// Should not panic and should return valid config
	config := rm.GetActiveConfiguration()
	if config == nil {
		t.Error("Expected non-nil configuration")
	}
}

func TestRunwayManager_ConfigIsCopy(t *testing.T) {
	runways := createTestRunways()
	rm := NewRunwayManager(runways, nil)

	// Get configuration
	config1 := rm.GetActiveConfiguration()

	// Modify the returned config
	delete(config1, "09L")

	// Get configuration again
	config2 := rm.GetActiveConfiguration()

	// Should still have all runways (returned copy wasn't internal state)
	if len(config2) != 3 {
		t.Errorf("Expected 3 runways in new config, got %d - config was not a copy", len(config2))
	}

	if _, exists := config2["09L"]; !exists {
		t.Error("09L should still exist - external modification affected internal state")
	}
}
