package core

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTaskQueue(t *testing.T) {
	tq := NewPriorityQueue[int]()

	require.True(t, tq.Push("one", 1, 0))
	require.False(t, tq.Push("one", 1, 0))

	require.Equal(t, 1, tq.Len())

	val, ok := tq.Find("one")
	require.Equal(t, 1, val)
	require.True(t, ok)

	val, ok = tq.Peek()
	require.True(t, ok)
	require.Equal(t, 1, val)
	require.Equal(t, 1, tq.Len())

	val, ok = tq.Pop()
	require.Equal(t, 1, val)
	require.True(t, ok)
	require.Equal(t, 0, tq.Len())

	_, ok = tq.Pop()
	require.False(t, ok)
	require.False(t, tq.Drop("one"))
	require.True(t, tq.Push("one", 1, 0))
	require.True(t, tq.Drop("one"))
	require.Equal(t, 0, tq.Len())

	require.True(t, tq.Push("one", 1, 0))
	require.True(t, tq.Push("two", 2, 0))
	require.True(t, tq.Push("three", 3, 0))
	require.True(t, tq.Push("four", 4, 0))
	require.Equal(t, 4, tq.Len())

	require.True(t, tq.UpdatePriority("three", 2))
	require.True(t, tq.UpdatePriority("one", 0))
	require.True(t, tq.UpdatePriority("two", 1))
	require.True(t, tq.Drop("four"))
	require.False(t, tq.UpdatePriority("five", 1))
	require.Equal(t, 3, tq.Len())
	require.Len(t, tq.All(), 3)

	require.False(t, tq.Push("two", 2, 10))

	val, ok = tq.Peek()
	require.True(t, ok)
	require.Equal(t, 3, val)
	require.Equal(t, 3, tq.Len())

	val, ok = tq.Pop()
	require.True(t, ok)
	require.Equal(t, 3, val)
	require.Equal(t, 2, tq.Len())

	val, ok = tq.Pop()
	require.True(t, ok)
	require.Equal(t, 2, val)
	require.Equal(t, 1, tq.Len())

	val, ok = tq.Pop()
	require.True(t, ok)
	require.Equal(t, 1, val)
	require.Equal(t, 0, tq.Len())

	_, ok = tq.Peek()
	require.False(t, ok)

	_, ok = tq.Find("one")
	require.False(t, ok)
}
