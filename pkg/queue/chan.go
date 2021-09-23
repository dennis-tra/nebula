package queue

type FIFO struct {
	buf []interface{}
	in  chan interface{}
	out chan interface{}
}

func NewFIFO() *FIFO {
	fifo := &FIFO{
		buf: []interface{}{}, // TODO: use a ring buffer?
		in:  make(chan interface{}),
		out: make(chan interface{}),
	}

	go fifo.listen()

	return fifo
}

func (fifo *FIFO) DoneProducing() {
	close(fifo.in)
}

func (fifo *FIFO) Produce() chan<- interface{} {
	return fifo.in
}

func (fifo *FIFO) Push(elem interface{}) {
	fifo.in <- elem
}

func (fifo *FIFO) Consume() <-chan interface{} {
	return fifo.out
}

func (fifo *FIFO) listen() {
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
