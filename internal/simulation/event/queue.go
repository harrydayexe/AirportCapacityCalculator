package event

import (
	"container/heap"
	"sync"
)

// EventQueue is a priority queue of events ordered by time.
// Events are processed chronologically from earliest to latest.
// This queue is safe for concurrent use by multiple goroutines.
type EventQueue struct {
	items *eventHeap
	mu    sync.Mutex
}

// NewEventQueue creates a new empty event queue.
func NewEventQueue() *EventQueue {
	h := &eventHeap{}
	heap.Init(h)
	return &EventQueue{
		items: h,
	}
}

// Push adds an event to the queue.
// This method is safe for concurrent use.
func (q *EventQueue) Push(event Event) {
	q.mu.Lock()
	defer q.mu.Unlock()
	heap.Push(q.items, event)
}

// Pop removes and returns the earliest event from the queue.
// Returns nil if the queue is empty.
// This method is safe for concurrent use.
func (q *EventQueue) Pop() Event {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.items.Len() == 0 {
		return nil
	}
	return heap.Pop(q.items).(Event)
}

// Peek returns the earliest event without removing it.
// Returns nil if the queue is empty.
// This method is safe for concurrent use.
func (q *EventQueue) Peek() Event {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.items.Len() == 0 {
		return nil
	}
	return (*q.items)[0]
}

// Len returns the number of events in the queue.
// This method is safe for concurrent use.
func (q *EventQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.items.Len()
}

// HasNext returns true if there are more events in the queue.
// This method is safe for concurrent use.
func (q *EventQueue) HasNext() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.items.Len() > 0
}

// eventHeap implements heap.Interface for Event items ordered by time.
type eventHeap []Event

func (h eventHeap) Len() int {
	return len(h)
}

func (h eventHeap) Less(i, j int) bool {
	// Earlier events have higher priority
	return h[i].Time().Before(h[j].Time())
}

func (h eventHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *eventHeap) Push(x any) {
	*h = append(*h, x.(Event))
}

func (h *eventHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}
