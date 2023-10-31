package core

import (
	"container/heap"
)

// An Task is something we manage in a Priority queue.
type Task[T any] struct {
	Value    T   // The value of the item; arbitrary.
	Priority int // The priority of the item in the queue.
	Index    int // The index of the item in the heap.
}

// A PriorityQueue implements heap.Interface and holds Tasks.
type PriorityQueue[T any] []*Task[T]

func (pq PriorityQueue[T]) Len() int { return len(pq) }

func (pq PriorityQueue[T]) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].Priority > pq[j].Priority
}

func (pq PriorityQueue[T]) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *PriorityQueue[T]) Push(x any) {
	n := len(*pq)
	item := x.(*Task[T])
	item.Index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue[T]) Pop() any {
	return any(pq.PopTyped())
}

func (pq *PriorityQueue[T]) PopTyped() *Task[T] {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.Index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

func (pq *PriorityQueue[T]) Peek() *Task[T] {
	return (*pq)[0]
}

// update modifies the Priority and value of an Task in the queue.
func (pq *PriorityQueue[T]) update(item *Task[T], value T, priority int) {
	item.Value = value
	item.Priority = priority
	heap.Fix(pq, item.Index)
}

// This example creates a PriorityQueue with some items, adds and manipulates an item,
// and then removes the items in Priority order.
//func main() {
//	// Some items and their priorities.
//	items := map[string]int{
//		"banana": 3, "apple": 2, "pear": 4,
//	}
//
//	// Create a priority queue, put the items in it, and
//	// establish the priority queue (heap) invariants.
//	pq := make(PriorityQueue, len(items))
//	i := 0
//	for value, priority := range items {
//		pq[i] = &Task{
//			Value:    value,
//			Priority: priority,
//			Index:    i,
//		}
//		i++
//	}
//	heap.Init(&pq)
//
//	// Insert a new item and then modify its priority.
//	item := &Task{
//		Value:    "orange",
//		Priority: 1,
//	}
//	heap.Push(&pq, item)
//	pq.update(item, item.Value, 5)
//
//	// Take the items out; they arrive in decreasing priority order.
//	for pq.Len() > 0 {
//		item := heap.Pop(&pq).(*Task)
//		fmt.Printf("%.2d:%s ", item.Priority, item.Value)
//	}
//}
