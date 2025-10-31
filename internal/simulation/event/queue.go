package event

import (
	"container/heap"
)

// EventQueue is a priority queue of events ordered by time.
// Events are processed chronologically from earliest to latest.
type EventQueue struct {
	items *eventHeap
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
func (q *EventQueue) Push(event Event) {
	heap.Push(q.items, event)
}

// Pop removes and returns the earliest event from the queue.
// Returns nil if the queue is empty.
func (q *EventQueue) Pop() Event {
	if q.items.Len() == 0 {
		return nil
	}
	return heap.Pop(q.items).(Event)
}

// Peek returns the earliest event without removing it.
// Returns nil if the queue is empty.
func (q *EventQueue) Peek() Event {
	if q.items.Len() == 0 {
		return nil
	}
	return (*q.items)[0]
}

// Len returns the number of events in the queue.
func (q *EventQueue) Len() int {
	return q.items.Len()
}

// HasNext returns true if there are more events in the queue.
func (q *EventQueue) HasNext() bool {
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

func (h *eventHeap) Push(x interface{}) {
	*h = append(*h, x.(Event))
}

func (h *eventHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}
