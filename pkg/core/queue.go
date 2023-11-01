package core

import (
	"container/heap"

	log "github.com/sirupsen/logrus"
)

// An item is something we manage in a priority queue.
type item[T any] struct {
	value    T // The value of the item; arbitrary.
	key      string
	priority int // The priority of the item in the queue.
	index    int // The index of the item in the heap.
}

type PriorityQueue[T any] struct {
	queue  *priorityQueue[T]
	lookup map[string]*item[T]
}

func NewPriorityQueue[T any]() *PriorityQueue[T] {
	queue := make(priorityQueue[T], 0)
	return &PriorityQueue[T]{
		queue:  &queue,
		lookup: map[string]*item[T]{},
	}
}

func (tq *PriorityQueue[T]) Push(key string, value T, priority int) bool {
	if _, found := tq.lookup[key]; found {
		return false
	}
	item := &item[T]{
		value:    value,
		key:      key,
		priority: priority,
	}
	heap.Push(tq.queue, item)
	tq.lookup[item.key] = item
	return true
}

func (tq *PriorityQueue[T]) Find(key string) (T, bool) {
	item, found := tq.lookup[key]
	if !found {
		return *new(T), false
	}
	return item.value, true
}

func (tq *PriorityQueue[T]) Peek() (T, bool) {
	if tq.queue.Len() == 0 {
		return *new(T), false
	}
	return (*tq.queue)[0].value, true
}

func (tq *PriorityQueue[T]) Pop() (T, bool) {
	if tq.queue.Len() == 0 {
		return *new(T), false
	}
	item := heap.Pop(tq.queue).(*item[T])
	delete(tq.lookup, item.key)
	return item.value, true
}

func (tq *PriorityQueue[T]) Drop(key string) bool {
	item, found := tq.lookup[key]
	if !found {
		return false
	}
	heap.Remove(tq.queue, item.index)
	delete(tq.lookup, item.key)
	return true
}

// UpdatePriority modifies the priority of an item in the queue.
func (tq *PriorityQueue[T]) UpdatePriority(key string, priority int) bool {
	item, found := tq.lookup[key]
	if !found {
		return false
	} else if item.priority == priority {
		return true
	}
	item.priority = priority
	heap.Fix(tq.queue, item.index)
	return true
}

// Update modifies the priority and value of an item in the queue.
func (tq *PriorityQueue[T]) Update(key string, value T, priority int) bool {
	item, found := tq.lookup[key]
	if !found {
		return false
	}
	if item.priority < priority {
		log.Error("FIXEd")
	}
	item.priority = priority
	item.value = value
	heap.Fix(tq.queue, item.index)
	return true
}

func (tq *PriorityQueue[T]) Len() int {
	return tq.queue.Len()
}

func (tq *PriorityQueue[T]) All() map[string]T {
	out := make(map[string]T, len(tq.lookup))
	for key, item := range tq.lookup {
		out[key] = item.value
	}
	return out
}

// A priorityQueue implements heap.Interface and holds Items.
type priorityQueue[T any] []*item[T]

func (pq priorityQueue[T]) Len() int { return len(pq) }

func (pq priorityQueue[T]) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].priority > pq[j].priority
}

func (pq priorityQueue[T]) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue[T]) Push(x any) {
	n := len(*pq)
	item := x.(*item[T])
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue[T]) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}
