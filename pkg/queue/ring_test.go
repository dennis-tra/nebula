package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ring(t *testing.T) {
	size := int64(2)

	ring := NewRing(size)
	assert.EqualValues(t, 0, ring.tail)
	assert.EqualValues(t, 0, ring.head)
	assert.EqualValues(t, 0, ring.len)
	assert.EqualValues(t, size, ring.mod)
	assert.EqualValues(t, size, len(ring.buffer))

	ring.Push(1)
	assert.EqualValues(t, 1, ring.tail)
	assert.EqualValues(t, 0, ring.head)
	assert.EqualValues(t, 1, ring.len)
	assert.EqualValues(t, size, ring.mod)
	assert.EqualValues(t, size, len(ring.buffer))

	ring.Push(2)
	assert.EqualValues(t, 2, ring.tail)
	assert.EqualValues(t, 0, ring.head)
	assert.EqualValues(t, 2, ring.len)
	assert.EqualValues(t, size*2, ring.mod)
	assert.EqualValues(t, size*2, len(ring.buffer))

	val := ring.Peek()
	assert.EqualValues(t, 1, val)

	val, ok := ring.Pop()
	assert.True(t, ok)

	assert.EqualValues(t, 1, val)
	assert.EqualValues(t, 2, ring.tail)
	assert.EqualValues(t, 1, ring.head)
	assert.EqualValues(t, 1, ring.len)
	assert.EqualValues(t, size*2, ring.mod)
	assert.EqualValues(t, size*2, len(ring.buffer))

	assert.False(t, ring.Empty())

	val, ok = ring.Pop()
	assert.True(t, ok)

	assert.EqualValues(t, 2, val)
	assert.EqualValues(t, 2, ring.tail)
	assert.EqualValues(t, 2, ring.head)
	assert.EqualValues(t, 0, ring.len)
	assert.EqualValues(t, size*2, ring.mod)
	assert.EqualValues(t, size*2, len(ring.buffer))

	assert.True(t, ring.Empty())

	ring.Push(3)
	ring.Push(4)

	val, ok = ring.Pop()
	assert.True(t, ok)

	assert.EqualValues(t, 3, val)
	assert.EqualValues(t, 0, ring.tail)
	assert.EqualValues(t, 3, ring.head)
	assert.EqualValues(t, 1, ring.len)
	assert.EqualValues(t, size*2, ring.mod)
	assert.EqualValues(t, size*2, len(ring.buffer))
}
