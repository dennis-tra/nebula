package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFIFO(t *testing.T) {
	fifo := NewFIFO()
	assert.NotNil(t, fifo)
}

func TestFIFO_Close(t *testing.T) {
	fifo := NewFIFO()
	fifo.DoneProducing()
}

func TestFIFO_Push(t *testing.T) {
	fifo := NewFIFO()
	fifo.Push(1)
	fifo.Push(2)
	assert.Equal(t, 1, fifo.Pop())
	assert.Equal(t, 2, fifo.Pop())
	fifo.DoneProducing()
}

func TestFIFO_Mixed(t *testing.T) {
	fifo := NewFIFO()
	assert.True(t, fifo.Push(1))
	fifo.DoneProducing()
	assert.False(t, fifo.Push(2))
	assert.Equal(t, 1, fifo.Pop())
	fifo.DoneConsuming()
	err := fifo.Pop().(error)
	assert.Error(t, err)
}

func TestFIFO_Push_Done(t *testing.T) {
	fifo := NewFIFO()
	fifo.Push(1)
	fifo.Push(2)
	assert.Equal(t, 1, fifo.Pop())
	assert.Equal(t, 2, fifo.Pop())
	fifo.DoneProducing()
}

func TestFIFO_Push_Done2(t *testing.T) {
	fifo := NewFIFO()
	assert.True(t, fifo.Push(1))
	assert.True(t, fifo.Push(2))
	fifo.DoneProducing()
	assert.False(t, fifo.Push(3))
}

func TestFIFO_Async(t *testing.T) {
	fifo := NewFIFO()
	go func() {
		for i := 0; i < 100; i++ {
			fifo.Push(i)
		}
		fifo.DoneProducing()
	}()

	i := 0
	for range fifo.Consume() {
		i += 1
	}
	assert.Equal(t, 100, i)
}
