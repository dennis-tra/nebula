package queue

type Ring[T any] struct {
	buffer []*T
	head   int64
	tail   int64
	mod    int64
	len    int64
}

func NewRing[T any](size int64) *Ring[T] {
	return &Ring[T]{
		buffer: make([]*T, size),
		head:   0,
		tail:   0,
		mod:    size,
		len:    0,
	}
}

func (q *Ring[T]) Push(item T) {
	q.tail = (q.tail + 1) % q.mod
	if q.tail == q.head {
		// we need to resize
		newLen := q.mod * 2
		newBuff := make([]*T, newLen)
		for i := int64(0); i < q.mod; i++ {
			buffIndex := (q.tail + i) % q.mod
			newBuff[i] = q.buffer[buffIndex]
		}
		// set the new buffer and reset head and tail
		q.buffer = newBuff
		q.head = 0
		q.tail = q.mod
		q.mod = newLen
	}
	q.len += 1
	q.buffer[q.tail] = &item
}

func (q *Ring[T]) Length() int64 {
	return q.len
}

func (q *Ring[T]) Empty() bool {
	return q.Length() == 0
}

func (q *Ring[T]) Pop() (T, bool) {
	if q.Empty() {
		return *new(T), false
	}
	q.head = (q.head + 1) % q.mod
	res := q.buffer[q.head]
	q.buffer[q.head] = nil
	q.len -= 1
	return *res, true
}

func (q *Ring[T]) Peek() T {
	return *q.buffer[(q.head+1)%q.mod]
}

func (q *Ring[T]) PopMany(count int64) []T {
	if count >= q.len {
		count = q.len
	}

	q.len -= count
	buffer := make([]T, count)
	for i := int64(0); i < count; i++ {
		pos := (q.head + 1 + i) % q.mod
		buffer[i] = *q.buffer[pos]
		q.buffer[pos] = nil
	}
	q.head = (q.head + count) % q.mod

	return buffer
}
