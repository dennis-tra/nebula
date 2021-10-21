package queue

import (
	"sync"
	"sync/atomic"
)

type buffer struct {
	buffer []interface{}
	head   int64
	tail   int64
	mod    int64
}

type ring struct {
	len     int64
	content *buffer
	lock    sync.Mutex
}

func New(initialSize int64) *ring {
	return &ring{
		content: &buffer{
			buffer: make([]interface{}, initialSize),
			head:   0,
			tail:   0,
			mod:    initialSize,
		},
		len: 0,
	}
}

func (q *ring) Push(item interface{}) {
	q.lock.Lock()
	c := q.content
	c.tail = (c.tail + 1) % c.mod
	if c.tail == c.head {
		var fillFactor int64 = 2
		// we need to resize

		newLen := c.mod * fillFactor
		newBuff := make([]interface{}, newLen)

		for i := int64(0); i < c.mod; i++ {
			buffIndex := (c.tail + i) % c.mod
			newBuff[i] = c.buffer[buffIndex]
		}
		// set the new buffer and reset head and tail
		newContent := &buffer{
			buffer: newBuff,
			head:   0,
			tail:   c.mod,
			mod:    newLen,
		}
		q.content = newContent
	}
	atomic.AddInt64(&q.len, 1)
	q.content.buffer[q.content.tail] = item
	q.lock.Unlock()
}

func (q *ring) Length() int64 {
	return atomic.LoadInt64(&q.len)
}

func (q *ring) Empty() bool {
	return q.Length() == 0
}

// single consumer
func (q *ring) Pop() (interface{}, bool) {
	if q.Empty() {
		return nil, false
	}
	// as we are a single consumer, no other thread can have poped the items there are guaranteed to be items now

	q.lock.Lock()
	c := q.content
	c.head = (c.head + 1) % c.mod
	res := c.buffer[c.head]
	c.buffer[c.head] = nil
	atomic.AddInt64(&q.len, -1)
	q.lock.Unlock()
	return res, true
}

func (q *ring) Peek() interface{} {
	if q.Empty() {
		return nil
	}
	// as we are a single consumer, no other thread can have poped the items there are guaranteed to be items now

	q.lock.Lock()
	c := q.content
	res := c.buffer[(c.head+1)%c.mod]
	q.lock.Unlock()
	return res
}

func (q *ring) PopMany(count int64) []interface{} {
	if q.Empty() {
		return nil
	}

	q.lock.Lock()
	c := q.content

	if count >= q.len {
		count = q.len
	}
	atomic.AddInt64(&q.len, -count)

	buffer := make([]interface{}, count)
	for i := int64(0); i < count; i++ {
		pos := (c.head + 1 + i) % c.mod
		buffer[i] = c.buffer[pos]
		c.buffer[pos] = nil
	}
	c.head = (c.head + count) % c.mod

	q.lock.Unlock()
	return buffer
}
