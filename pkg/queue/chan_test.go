package queue

import (
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FIFO(t *testing.T) {
	fifo := NewFIFO[int]()
	for i := 0; i < 10; i++ {
		fifo.Push(i)
	}
	for i := 0; i < 10; i++ {
		assert.Equal(t, i, <-fifo.Consume())
	}
	for i := 0; i < 10; i++ {
		fifo.Push(i)
	}
	fifo.DoneProducing()
	for i := 0; i < 10; i++ {
		assert.Equal(t, i, <-fifo.Consume())
	}
}

func Test_FIFO_async(t *testing.T) {
	fifo := NewFIFO[int]()

	var wg sync.WaitGroup
	wg.Add(10)
	go func() {
		for i := 0; i < 10; i++ {
			assert.Equal(t, i, <-fifo.Consume())
			wg.Done()
		}
	}()

	runtime.Gosched()

	for i := 0; i < 10; i++ {
		fifo.Push(i)
	}

	wg.Wait()

	fifo.Push(1)
	fifo.DoneProducing()
	assert.Equal(t, 1, <-fifo.Consume())
}
