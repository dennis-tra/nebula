package core

import (
	"context"
	"sync"
)

type Worker[T any, R any] interface {
	Work(ctx context.Context, task T) (R, error)
}

type Pool[T any, R any] struct {
	start   sync.Once
	results chan Result[R]
	workers []Worker[T, R]
}

func NewPool[T any, R any](workers ...Worker[T, R]) *Pool[T, R] {
	return &Pool[T, R]{
		results: make(chan Result[R]),
		workers: workers,
	}
}

func (w *Pool[T, R]) Start(ctx context.Context, tasks <-chan T) <-chan Result[R] {
	w.start.Do(func() {
		var wg sync.WaitGroup
		for _, worker := range w.workers {
			wg.Add(1)
			worker := worker
			go func() {
				for task := range tasks {
					result, err := worker.Work(ctx, task)
					w.results <- Result[R]{
						Value: result,
						Error: err,
					}
				}
				wg.Done()
			}()
		}

		go func() {
			wg.Wait()
			close(w.results)
		}()
	})

	return w.results
}

func (w *Pool[T, R]) Size() int {
	return len(w.workers)
}

type Result[R any] struct {
	Value R
	Error error
}
