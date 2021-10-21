package queue

type ring struct {
	buffer []interface{}
	head   int64
	tail   int64
	mod    int64
	len    int64
}

func New(size int64) *ring {
	return &ring{
		buffer: make([]interface{}, size),
		head:   0,
		tail:   0,
		mod:    size,
		len:    0,
	}
}

func (q *ring) Push(item interface{}) {
	q.tail = (q.tail + 1) % q.mod
	if q.tail == q.head {
		// we need to resize
		newLen := q.mod * 2
		newBuff := make([]interface{}, newLen)
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
	q.buffer[q.tail] = item
}

func (q *ring) Length() int64 {
	return q.len
}

func (q *ring) Empty() bool {
	return q.Length() == 0
}

func (q *ring) Pop() (interface{}, bool) {
	if q.Empty() {
		return nil, false
	}
	q.head = (q.head + 1) % q.mod
	res := q.buffer[q.head]
	q.buffer[q.head] = nil
	q.len -= 1
	return res, true
}

func (q *ring) Peek() interface{} {
	return q.buffer[(q.head+1)%q.mod]
}

func (q *ring) PopMany(count int64) []interface{} {
	if count >= q.len {
		count = q.len
	}

	q.len -= count
	buffer := make([]interface{}, count)
	for i := int64(0); i < count; i++ {
		pos := (q.head + 1 + i) % q.mod
		buffer[i] = q.buffer[pos]
		q.buffer[pos] = nil
	}
	q.head = (q.head + count) % q.mod

	return buffer
}
