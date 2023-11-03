package core

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

var goleakIgnore = []goleak.Option{
	// https://github.com/census-instrumentation/opencensus-go/issues/1191
	goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	// could be solved with ipfslog.WriterGroup.Close()
	goleak.IgnoreTopFunction("github.com/ipfs/go-log/writer.(*MirrorWriter).logRoutine"),
}

func TestPoolNoWorker(t *testing.T) {
	tasks := make(chan string)
	pool := NewPool[string, int]()
	pool.Start(context.Background(), tasks)
	assert.Equal(t, 0, pool.Size())
	close(tasks)
}

func TestPoolSingleWorker(t *testing.T) {
	t.Run("no work", func(t *testing.T) {
		defer goleak.VerifyNone(t, goleakIgnore...)
		tasks := make(chan string)
		pool := NewPool[string, int](newTestWorker[string, int]())
		assert.Equal(t, 1, pool.Size())
		results := pool.Start(context.Background(), tasks)
		close(tasks)
		<-results
	})

	t.Run("performs work", func(t *testing.T) {
		defer goleak.VerifyNone(t, goleakIgnore...)
		ctx := context.Background()
		worker := newTestWorker[string, int]()
		worker.On("Work", mock.Anything, "4").Return(4, nil)
		pool := NewPool[string, int](worker)
		tasks := make(chan string)
		results := pool.Start(ctx, tasks)
		tasks <- "4"
		close(tasks)
		_, more := <-results
		assert.True(t, more)
		_, more = <-results
		assert.False(t, more)
	})

	t.Run("multiple starts", func(t *testing.T) {
		defer goleak.VerifyNone(t, goleakIgnore...)
		tasks := make(chan string)

		pool := NewPool[string, int](newAtoiWorker(t))
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
		defer goleak.VerifyNone(t, goleakIgnore...)
		ctx := context.Background()
		tasks := make(chan string)
		pool := NewPool[string, int](newAtoiWorker(t), newAtoiWorker(t))
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
		defer goleak.VerifyNone(t, goleakIgnore...)
		ctx := context.Background()
		tasks := make(chan string)
		pool := NewPool[string, int](newAtoiWorker(t), newAtoiWorker(t), newAtoiWorker(t))
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
		defer goleak.VerifyNone(t, goleakIgnore...)
		ctx := context.Background()
		tasks := make(chan string)

		workerCount := 3

		hooks := make([]chan struct{}, workerCount)
		workers := make([]*testWorker[string, int], workerCount)
		for i := 0; i < workerCount; i++ {
			hook := make(chan struct{})
			hooks[i] = hook

			worker := newTestWorker[string, int]()
			workerCall := worker.On("Work", mock.Anything, mock.Anything)
			workerCall.RunFn = func(args mock.Arguments) {
				<-hook
				val, err := strconv.Atoi(args.String(1))
				require.NoError(t, err)
				workerCall.ReturnArguments = mock.Arguments{val, nil}
			}
			workers[i] = worker
		}

		pool := NewPool[string, int](workers[0], workers[1], workers[2])
		results := pool.Start(ctx, tasks)
		for i := 0; i < workerCount; i++ {
			tasks <- strconv.Itoa(i)
		}
		close(tasks)
		for i := 0; i < workerCount; i++ {
			close(hooks[i])
			_, more := <-results
			assert.True(t, more)
		}

		_, more := <-results
		assert.False(t, more)
	})
}
