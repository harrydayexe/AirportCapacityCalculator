package event

import (
	"context"
	"testing"
	"time"
)

// mockWorldState for testing wind events
type mockWindWorldState struct {
	windSpeed     float64
	windDirection float64
	setWindCalled bool
	setWindError  error
}

func (m *mockWindWorldState) SetWind(speed, direction float64) error {
	m.windSpeed = speed
	m.windDirection = direction
	m.setWindCalled = true
	return m.setWindError
}

func (m *mockWindWorldState) GetWindSpeed() float64              { return m.windSpeed }
func (m *mockWindWorldState) GetWindDirection() float64          { return m.windDirection }
func (m *mockWindWorldState) SetCurfewActive(active bool)        {}
func (m *mockWindWorldState) GetCurfewActive() bool              { return false }
func (m *mockWindWorldState) SetRunwayAvailable(id string, a bool) error { return nil }
func (m *mockWindWorldState) GetRunwayAvailable(id string) (bool, error) { return true, nil }
func (m *mockWindWorldState) SetRotationMultiplier(multiplier float32)    {}
func (m *mockWindWorldState) GetRotationMultiplier() float32     { return 1.0 }
func (m *mockWindWorldState) SetGateCapacityConstraint(constraint float32) error { return nil }
func (m *mockWindWorldState) GetGateCapacityConstraint() float32 { return 0 }
func (m *mockWindWorldState) SetTaxiTimeOverhead(d time.Duration) error { return nil }
func (m *mockWindWorldState) GetTaxiTimeOverhead() time.Duration { return 0 }
func (m *mockWindWorldState) SetActiveRunwayConfiguration(c map[string]*ActiveRunwayInfo) error {
	return nil
}
func (m *mockWindWorldState) GetActiveRunwayConfiguration() map[string]*ActiveRunwayInfo {
	return nil
}
func (m *mockWindWorldState) NotifyRunwayAvailabilityChange(id string, a bool, t time.Time) error {
	return nil
}
func (m *mockWindWorldState) NotifyCurfewChange(a bool, t time.Time) error {
	return nil
}

// TestNewWindChangeEvent tests the constructor
func TestNewWindChangeEvent(t *testing.T) {
	timestamp := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	event := NewWindChangeEvent(15.0, 270.0, timestamp)

	if event == nil {
		t.Fatal("Expected non-nil event")
	}

	if event.GetSpeed() != 15.0 {
		t.Errorf("Expected speed 15.0, got %f", event.GetSpeed())
	}

	if event.GetDirection() != 270.0 {
		t.Errorf("Expected direction 270.0, got %f", event.GetDirection())
	}

	if !event.Time().Equal(timestamp) {
		t.Errorf("Expected timestamp %v, got %v", timestamp, event.Time())
	}
}

// TestWindChangeEventTime tests the Time method
func TestWindChangeEventTime(t *testing.T) {
	timestamp := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)
	event := NewWindChangeEvent(20.0, 180.0, timestamp)

	if !event.Time().Equal(timestamp) {
		t.Errorf("Expected time %v, got %v", timestamp, event.Time())
	}
}

// TestWindChangeEventType tests the Type method
func TestWindChangeEventType(t *testing.T) {
	event := NewWindChangeEvent(10.0, 90.0, time.Now())

	if event.Type() != WindChangeType {
		t.Errorf("Expected type %v, got %v", WindChangeType, event.Type())
	}

	if event.Type().String() != "WindChange" {
		t.Errorf("Expected type string 'WindChange', got '%s'", event.Type().String())
	}
}

// TestWindChangeEventApply tests the Apply method
func TestWindChangeEventApply(t *testing.T) {
	tests := []struct {
		name      string
		speed     float64
		direction float64
	}{
		{"Calm wind", 0, 0},
		{"Light easterly", 10, 90},
		{"Moderate westerly", 20, 270},
		{"Strong northerly", 35, 360},
		{"Southerly", 15, 180},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := NewWindChangeEvent(tt.speed, tt.direction, time.Now())
			mockWorld := &mockWindWorldState{}

			err := event.Apply(context.Background(), mockWorld)

			if err != nil {
				t.Errorf("Apply returned unexpected error: %v", err)
			}

			if !mockWorld.setWindCalled {
				t.Error("SetWind was not called")
			}

			if mockWorld.windSpeed != tt.speed {
				t.Errorf("Expected wind speed %f, got %f", tt.speed, mockWorld.windSpeed)
			}

			if mockWorld.windDirection != tt.direction {
				t.Errorf("Expected wind direction %f, got %f", tt.direction, mockWorld.windDirection)
			}
		})
	}
}

// TestWindChangeEventGetters tests the getter methods
func TestWindChangeEventGetters(t *testing.T) {
	event := NewWindChangeEvent(25.5, 123.4, time.Now())

	if event.GetSpeed() != 25.5 {
		t.Errorf("GetSpeed: expected 25.5, got %f", event.GetSpeed())
	}

	if event.GetDirection() != 123.4 {
		t.Errorf("GetDirection: expected 123.4, got %f", event.GetDirection())
	}
}

// TestWindChangeEventMultipleChanges tests a sequence of wind changes
func TestWindChangeEventMultipleChanges(t *testing.T) {
	mockWorld := &mockWindWorldState{}
	ctx := context.Background()

	// Morning: calm
	event1 := NewWindChangeEvent(0, 0, time.Date(2024, 1, 1, 6, 0, 0, 0, time.UTC))
	if err := event1.Apply(ctx, mockWorld); err != nil {
		t.Fatalf("Event 1 failed: %v", err)
	}
	if mockWorld.windSpeed != 0 || mockWorld.windDirection != 0 {
		t.Error("Morning wind incorrect")
	}

	// Noon: westerly builds
	event2 := NewWindChangeEvent(15, 270, time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC))
	if err := event2.Apply(ctx, mockWorld); err != nil {
		t.Fatalf("Event 2 failed: %v", err)
	}
	if mockWorld.windSpeed != 15 || mockWorld.windDirection != 270 {
		t.Error("Noon wind incorrect")
	}

	// Evening: strong westerly
	event3 := NewWindChangeEvent(25, 270, time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC))
	if err := event3.Apply(ctx, mockWorld); err != nil {
		t.Fatalf("Event 3 failed: %v", err)
	}
	if mockWorld.windSpeed != 25 || mockWorld.windDirection != 270 {
		t.Error("Evening wind incorrect")
	}
}
