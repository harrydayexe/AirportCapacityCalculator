package policy

import (
	"io"
	"log/slog"
	"time"

	"github.com/harrydayexe/AirportCapacityCalculator/internal/simulation/event"
)

// testLogger creates a test logger that discards output
func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// mockEventWorld provides a test implementation of the EventWorld interface
type mockEventWorld struct {
	startTime time.Time
	endTime   time.Time
	runwayIDs []string
	events    []event.Event
}

// newMockEventWorld creates a new mock event world
func newMockEventWorld(startTime, endTime time.Time, runwayIDs []string) *mockEventWorld {
	return &mockEventWorld{
		startTime: startTime,
		endTime:   endTime,
		runwayIDs: runwayIDs,
		events:    []event.Event{},
	}
}

func (m *mockEventWorld) ScheduleEvent(evt event.Event) {
	m.events = append(m.events, evt)
}

func (m *mockEventWorld) GetEventQueue() *event.EventQueue {
	queue := event.NewEventQueue()
	for _, evt := range m.events {
		queue.Push(evt)
	}
	return queue
}

func (m *mockEventWorld) GetStartTime() time.Time {
	return m.startTime
}

func (m *mockEventWorld) GetEndTime() time.Time {
	return m.endTime
}

func (m *mockEventWorld) GetRunwayIDs() []string {
	return m.runwayIDs
}

// Helper to count events by type
func (m *mockEventWorld) CountEventsByType(eventType event.EventType) int {
	count := 0
	for _, evt := range m.events {
		if evt.Type() == eventType {
			count++
		}
	}
	return count
}

// Helper to get all events
func (m *mockEventWorld) GetEvents() []event.Event {
	return m.events
}
