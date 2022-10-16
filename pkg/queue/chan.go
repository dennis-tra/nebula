package queue

type FIFO[T any] struct {
	ring *Ring[T]
	in   chan T
	out  chan T
}

func NewFIFO[T any]() *FIFO[T] {
	fifo := &FIFO[T]{
		ring: NewRing[T](1000),
		in:   make(chan T),
		out:  make(chan T),
	}

	go fifo.listen()

	return fifo
}

func (fifo *FIFO[T]) DoneProducing() {
	close(fifo.in)
}

func (fifo *FIFO[T]) Produce() chan<- T {
	return fifo.in
}

func (fifo *FIFO[T]) Push(elem T) {
	fifo.in <- elem
}

func (fifo *FIFO[T]) Consume() <-chan T {
	return fifo.out
}

func (fifo *FIFO[T]) listen() {
	defer close(fifo.out)

	var ok bool
	var elem T
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
		fifo.ring.Push(elem)

		for fifo.ring.Length() > 0 {
			select {
			case elem, ok := <-fifo.in:
				if !ok {
					// The sender has closed the channel
					break LOOP
				}
				fifo.ring.Push(elem)
			case fifo.out <- fifo.ring.Peek():
				fifo.ring.Pop()
			}
		}
	}

	for _, elem := range fifo.ring.PopMany(fifo.ring.Length()) {
		fifo.out <- elem
	}
}
