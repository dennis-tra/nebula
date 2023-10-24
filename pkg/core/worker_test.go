package core

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	nt "github.com/dennis-tra/nebula-crawler/pkg/nebtest"

	"github.com/stretchr/testify/assert"
)

func TestPoolNoWorker(t *testing.T) {
	tasks := make(chan string)
	pool := NewPool[string, int]()
	pool.Start(context.Background(), tasks)
	assert.Equal(t, 0, pool.Size())
	close(tasks)
}

func TestPoolSingleWorker(t *testing.T) {
	t.Run("no work", func(t *testing.T) {
		tasks := make(chan string)
		pool := NewPool[string, int](nt.NewWorker())
		assert.Equal(t, 1, pool.Size())
		results := pool.Start(context.Background(), tasks)
		close(tasks)
		<-results
	})

	t.Run("performs work", func(t *testing.T) {
		tasks := make(chan string)
		pool := NewPool[string, int](nt.NewWorker())
		results := pool.Start(context.Background(), tasks)
		tasks <- "4"
		close(tasks)
		<-results
		<-results
	})

	t.Run("multiple starts", func(t *testing.T) {
		tasks := make(chan string)
		pool := NewPool[string, int](nt.NewWorker())

		results := pool.Start(context.Background(), tasks)
		results = pool.Start(context.Background(), tasks)
		go func() { results = pool.Start(context.Background(), tasks) }()
		go func() { results = pool.Start(context.Background(), tasks) }()
		tasks <- "4"
		close(tasks)
		val, more := <-results
		assert.True(t, more)
		assert.Equal(t, 4, val.Value)
		_, more = <-results
		assert.False(t, more)
	})
}

func TestPoolMultipleWorkers(t *testing.T) {
	t.Run("equal workers and tasks", func(t *testing.T) {
		ctx := context.Background()
		tasks := make(chan string)
		pool := NewPool[string, int](nt.NewWorker(), nt.NewWorker())
		assert.Equal(t, 2, pool.Size())
		results := pool.Start(ctx, tasks)
		tasks <- "1"
		assert.Equal(t, 1, (<-results).Value)
		tasks <- "A"
		assert.Error(t, (<-results).Error)
		close(tasks)
		_, more := <-results
		assert.False(t, more)
	})

	t.Run("few workers and many tasks", func(t *testing.T) {
		ctx := context.Background()
		tasks := make(chan string)
		pool := NewPool[string, int](nt.NewWorker(), nt.NewWorker(), nt.NewWorker())
		assert.Equal(t, 3, pool.Size())
		results := pool.Start(ctx, tasks)

		go func() {
			for i := 0; i < 1000; i++ {
				tasks <- fmt.Sprint(i)
			}
			close(tasks)
		}()

		count := 0
		for result := range results {
			assert.NotNil(t, result.Value)
			assert.Nil(t, result.Error)
			count += 1
		}
		assert.Equal(t, 1000, count)
	})

	t.Run("work is distributed", func(t *testing.T) {
		ctx := context.Background()
		tasks := make(chan string)

		workerHook1 := make(chan struct{})
		workerHook2 := make(chan struct{})
		workerHook3 := make(chan struct{})
		pool := NewPool[string, int](
			&nt.Worker{
				WorkHook: func(ctx context.Context, task string) (int, error) {
					<-workerHook1
					return strconv.Atoi(task)
				},
			},
			&nt.Worker{
				WorkHook: func(ctx context.Context, task string) (int, error) {
					<-workerHook2
					return strconv.Atoi(task)
				},
			},
			&nt.Worker{
				WorkHook: func(ctx context.Context, task string) (int, error) {
					<-workerHook3
					return strconv.Atoi(task)
				},
			},
		)
		results := pool.Start(ctx, tasks)
		tasks <- "1"
		tasks <- "2"
		tasks <- "3"
		close(tasks)
		close(workerHook1)
		<-results
		close(workerHook2)
		<-results
		close(workerHook3)
		<-results

		_, more := <-results
		assert.False(t, more)
	})
}
