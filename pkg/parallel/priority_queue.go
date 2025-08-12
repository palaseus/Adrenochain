package parallel

import (
	"container/heap"
	"sync"
)

// PriorityQueueItem represents an item in the priority queue
type PriorityQueueItem struct {
	Item     *WorkItem
	Priority int
	Index    int
}

// PriorityQueue implements a thread-safe priority queue using a min-heap
type PriorityQueue struct {
	items []*PriorityQueueItem
	mu    sync.RWMutex
}

// NewPriorityQueue creates a new priority queue
func NewPriorityQueue() *PriorityQueue {
	pq := &PriorityQueue{
		items: make([]*PriorityQueueItem, 0),
	}
	heap.Init(pq)
	return pq
}

// PushItem adds an item to the priority queue
func (pq *PriorityQueue) PushItem(item *WorkItem, priority int) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	pqItem := &PriorityQueueItem{
		Item:     item,
		Priority: priority,
	}
	heap.Push(pq, pqItem)
}

// PopItem removes and returns the highest priority item
func (pq *PriorityQueue) PopItem() *WorkItem {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if pq.Len() == 0 {
		return nil
	}

	item := heap.Pop(pq).(*PriorityQueueItem)
	return item.Item
}

// Peek returns the highest priority item without removing it
func (pq *PriorityQueue) Peek() *WorkItem {
	pq.mu.RLock()
	defer pq.mu.RUnlock()

	if pq.Len() == 0 {
		return nil
	}

	return pq.items[0].Item
}

// Len returns the number of items in the queue
func (pq *PriorityQueue) Len() int {
	return len(pq.items)
}

// Empty returns true if the queue is empty
func (pq *PriorityQueue) Empty() bool {
	return pq.Len() == 0
}

// Clear removes all items from the queue
func (pq *PriorityQueue) Clear() {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	pq.items = make([]*PriorityQueueItem, 0)
	heap.Init(pq)
}

// heap.Interface implementation
func (pq *PriorityQueue) Less(i, j int) bool {
	// Lower priority number = higher priority
	return pq.items[i].Priority < pq.items[j].Priority
}

func (pq *PriorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].Index = i
	pq.items[j].Index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(pq.items)
	item := x.(*PriorityQueueItem)
	item.Index = n
	pq.items = append(pq.items, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := pq.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.Index = -1 // for safety
	pq.items = old[0 : n-1]
	return item
}
