package event

import (
	"context"
	"sync"
	"testing"
	"time"
)

// mockEvent is a simple event for testing
type mockEvent struct {
	timestamp time.Time
	eventType EventType
}

func (m *mockEvent) Time() time.Time {
	return m.timestamp
}

func (m *mockEvent) Type() EventType {
	return m.eventType
}

func (m *mockEvent) Apply(ctx context.Context, world WorldState) error {
	return nil
}

func TestEventQueue_ConcurrentPush(t *testing.T) {
	queue := NewEventQueue()
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Number of goroutines and events per goroutine
	numGoroutines := 10
	eventsPerGoroutine := 100
	totalEvents := numGoroutines * eventsPerGoroutine

	var wg sync.WaitGroup

	// Push events concurrently from multiple goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()
			for j := 0; j < eventsPerGoroutine; j++ {
				event := &mockEvent{
					timestamp: baseTime.Add(time.Duration(routineID*eventsPerGoroutine+j) * time.Second),
					eventType: CurfewStartType,
				}
				queue.Push(event)
			}
		}(i)
	}

	wg.Wait()

	// Verify all events were added
	if queue.Len() != totalEvents {
		t.Errorf("Expected %d events, got %d", totalEvents, queue.Len())
	}

	// Verify events are in chronological order
	var prevTime time.Time
	for queue.HasNext() {
		event := queue.Pop()
		if event == nil {
			t.Fatal("Got nil event from non-empty queue")
		}

		if !prevTime.IsZero() && event.Time().Before(prevTime) {
			t.Errorf("Events not in chronological order: %v came after %v", event.Time(), prevTime)
		}
		prevTime = event.Time()
	}
}

func TestEventQueue_ConcurrentPushAndPop(t *testing.T) {
	queue := NewEventQueue()
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	numPushers := 5
	numPoppers := 3
	eventsPerPusher := 100

	var wg sync.WaitGroup
	var pushersDone sync.WaitGroup

	// Pushers: add events concurrently
	pushersDone.Add(numPushers)
	for i := 0; i < numPushers; i++ {
		wg.Add(1)
		go func(pusherID int) {
			defer wg.Done()
			defer pushersDone.Done()
			for j := 0; j < eventsPerPusher; j++ {
				event := &mockEvent{
					timestamp: baseTime.Add(time.Duration(pusherID*eventsPerPusher+j) * time.Second),
					eventType: CurfewStartType,
				}
				queue.Push(event)
			}
		}(i)
	}

	// Poppers: remove events concurrently
	poppedCount := 0
	var poppedMu sync.Mutex
	var pushersFinished bool

	for i := 0; i < numPoppers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localCount := 0
			for {
				event := queue.Pop()
				if event == nil {
					// Queue is empty - check if pushers are done
					if pushersFinished && queue.Len() == 0 {
						break
					}
					time.Sleep(1 * time.Millisecond)
					continue
				}
				localCount++
			}

			poppedMu.Lock()
			poppedCount += localCount
			poppedMu.Unlock()
		}()
	}

	// Signal when all pushers are done
	go func() {
		pushersDone.Wait()
		pushersFinished = true
	}()

	wg.Wait()

	// Pop any remaining events
	for queue.HasNext() {
		if queue.Pop() != nil {
			poppedCount++
		}
	}

	// Verify we got all events
	totalEvents := numPushers * eventsPerPusher
	if poppedCount != totalEvents {
		t.Errorf("Expected %d events popped, got %d", totalEvents, poppedCount)
	}

	// Verify queue is empty
	if queue.Len() != 0 {
		t.Errorf("Expected empty queue, got %d events remaining", queue.Len())
	}
}

func TestEventQueue_ConcurrentLen(t *testing.T) {
	queue := NewEventQueue()
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	numGoroutines := 10
	eventsPerGoroutine := 50

	var wg sync.WaitGroup

	// Push events and check length concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()
			for j := 0; j < eventsPerGoroutine; j++ {
				event := &mockEvent{
					timestamp: baseTime.Add(time.Duration(routineID*eventsPerGoroutine+j) * time.Second),
					eventType: CurfewStartType,
				}
				queue.Push(event)

				// Concurrent Len() calls should not panic or corrupt state
				_ = queue.Len()
			}
		}(i)
	}

	wg.Wait()

	finalLen := queue.Len()
	expectedLen := numGoroutines * eventsPerGoroutine
	if finalLen != expectedLen {
		t.Errorf("Expected final length %d, got %d", expectedLen, finalLen)
	}
}

func TestEventQueue_ConcurrentPeek(t *testing.T) {
	queue := NewEventQueue()
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Add initial event
	firstEvent := &mockEvent{
		timestamp: baseTime,
		eventType: CurfewStartType,
	}
	queue.Push(firstEvent)

	var wg sync.WaitGroup
	numGoroutines := 10

	// Multiple goroutines peeking concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				peeked := queue.Peek()
				if peeked == nil {
					t.Error("Peek returned nil when queue should not be empty")
				} else if !peeked.Time().Equal(baseTime) {
					t.Errorf("Peek returned wrong event time: %v, expected %v", peeked.Time(), baseTime)
				}
			}
		}()
	}

	wg.Wait()

	// Queue should still have the event
	if queue.Len() != 1 {
		t.Errorf("Expected queue length 1, got %d", queue.Len())
	}
}

func TestEventQueue_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	queue := NewEventQueue()
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	numPushers := 20
	numPoppers := 10
	eventsPerPusher := 500

	var wg sync.WaitGroup

	// Pushers
	for i := 0; i < numPushers; i++ {
		wg.Add(1)
		go func(pusherID int) {
			defer wg.Done()
			for j := 0; j < eventsPerPusher; j++ {
				event := &mockEvent{
					timestamp: baseTime.Add(time.Duration(pusherID*eventsPerPusher+j) * time.Millisecond),
					eventType: EventType(j % 5), // Vary event types
				}
				queue.Push(event)
			}
		}(i)
	}

	// Poppers
	totalPopped := 0
	var poppedMu sync.Mutex

	for i := 0; i < numPoppers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localPopped := 0
			for {
				event := queue.Pop()
				if event == nil {
					time.Sleep(1 * time.Millisecond)
					if queue.Len() == 0 {
						break
					}
					continue
				}
				localPopped++
			}

			poppedMu.Lock()
			totalPopped += localPopped
			poppedMu.Unlock()
		}()
	}

	wg.Wait()

	// Pop any remaining
	for queue.HasNext() {
		if queue.Pop() != nil {
			totalPopped++
		}
	}

	expectedTotal := numPushers * eventsPerPusher
	if totalPopped != expectedTotal {
		t.Errorf("Expected %d total events popped, got %d", expectedTotal, totalPopped)
	}

	if queue.Len() != 0 {
		t.Errorf("Expected empty queue, got length %d", queue.Len())
	}
}
