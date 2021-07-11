package queue

import (
	"sync"
	"sync/atomic"
)

type ringBuffer struct {
	buffer []interface{}
	head   int64
	tail   int64
	mod    int64
}

type Queue struct {
	len      int64
	content  *ringBuffer
	items    chan interface{}
	listen   chan interface{}
	maxBatch int64
	lock     sync.Mutex
}

func New(initialSize int64) *Queue {
	q := &Queue{
		content: &ringBuffer{
			buffer: make([]interface{}, initialSize),
			head:   0,
			tail:   0,
			mod:    initialSize,
		},
		items:    make(chan interface{}),
		listen:   make(chan interface{}),
		maxBatch: 1000,
		len:      0,
	}

	go q.bind()

	return q
}

func (q *Queue) Close() {
	close(q.items)
	close(q.listen)
}

func (q *Queue) Listen() <-chan interface{} {
	return q.listen
}

func (q *Queue) bind() {
	for {
		batch := q.maxBatch
		if q.Length() < batch {
			batch = q.Length()
		}
		if got, ok := q.popMany(batch); ok {
			for _, item := range got {
				q.listen <- item
			}
		} else {
			item, ok := <-q.items
			if !ok {
				return
			}
			q.listen <- item
		}
	}
}

func (q *Queue) Push(item interface{}) {
	select {
	case q.items <- item:
		return
	default:
	}

	q.lock.Lock()
	defer q.lock.Unlock()

	q.content.tail = (q.content.tail + 1) % q.content.mod
	if q.content.tail == q.content.head {
		q.resize()
	}
	atomic.AddInt64(&q.len, 1)
	q.content.buffer[q.content.tail] = item
}

func (q *Queue) resize() {
	var fillFactor int64 = 2
	// we need to resize

	newLen := q.content.mod * fillFactor
	newBuff := make([]interface{}, newLen)

	for i := int64(0); i < q.content.mod; i++ {
		buffIndex := (q.content.tail + i) % q.content.mod
		newBuff[i] = q.content.buffer[buffIndex]
	}
	// set the new buffer and reset head and tail
	newContent := &ringBuffer{
		buffer: newBuff,
		head:   0,
		tail:   q.content.mod,
		mod:    newLen,
	}
	q.content = newContent
}

func (q *Queue) Length() int64 {
	return atomic.LoadInt64(&q.len)
}

func (q *Queue) Empty() bool {
	return q.Length() == 0
}

func (q *Queue) pop() (interface{}, bool) {
	if q.Empty() {
		return nil, false
	}
	// as we are a single consumer, no other thread can have popped the items there are guaranteed to be items now

	q.lock.Lock()
	c := q.content
	c.head = (c.head + 1) % c.mod
	res := c.buffer[c.head]
	c.buffer[c.head] = nil
	atomic.AddInt64(&q.len, -1)
	q.lock.Unlock()
	return res, true
}

func (q *Queue) popMany(count int64) ([]interface{}, bool) {
	if q.Empty() {
		return nil, false
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
	return buffer, true
}
