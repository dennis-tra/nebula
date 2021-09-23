package queue

type FIFO struct {
	buf      []interface{}
	in       chan interface{}
	out      chan interface{}
	shutdown chan struct{}
	done     chan struct{}
}

func NewFIFO() *FIFO {
	fifo := &FIFO{
		buf:      []interface{}{}, // TODO: use a ring buffer?
		in:       make(chan interface{}),
		out:      make(chan interface{}),
		shutdown: make(chan struct{}),
		done:     make(chan struct{}),
	}

	go fifo.listen()

	return fifo
}

// Done can be called by the user so that no more items can be pushed into this queue
func (fifo *FIFO) Done() {
	close(fifo.in)
}

// Close should be called by the user if he/she is done receiving elements
func (fifo *FIFO) Close() {
	close(fifo.shutdown)
	<-fifo.done
}

func (fifo *FIFO) Push(elem interface{}) {
	fifo.in <- elem
}

func (fifo *FIFO) Pop() interface{} {
	return <-fifo.out
}

func (fifo *FIFO) Listen() <-chan interface{} {
	return fifo.out
}

func (fifo *FIFO) listen() {
	defer close(fifo.done)
	defer close(fifo.out)

	var ok bool
	var elem interface{}
LOOP:
	for {

		// At the start the in channel is empty, so we're waiting for elements
		select {
		case elem, ok = <-fifo.in:
			if !ok {
				// The sender has closed the channel
				break LOOP
			}
		case <-fifo.shutdown:
			break LOOP
		}

		// Try to send new element immediately
		select {
		case fifo.out <- elem:
			continue
		default:
		}

		// We could not send the element, so we're buffering it
		fifo.buf = append(fifo.buf, elem)

		for len(fifo.buf) > 0 {
			select {
			case <-fifo.shutdown:
				break LOOP
			case elem, ok := <-fifo.in:
				if !ok {
					// The sender has closed the channel
					break LOOP
				}
				fifo.buf = append(fifo.buf, elem)
			case fifo.out <- fifo.buf[0]:
				fifo.buf = fifo.buf[1:]
			}
		}
	}
	for _, elem := range fifo.buf {
		fifo.out <- elem
	}
}
