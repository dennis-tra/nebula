package queue

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPushPop(t *testing.T) {
	q := New(10)
	q.Push("hello")
	res := <-q.Listen()
	assert.Equal(t, "hello", res)
	assert.True(t, q.Empty())
}

func TestPushPopRepeated(t *testing.T) {
	q := New(10)
	for i := 0; i < 100; i++ {
		q.Push("hello")
		res := <-q.Listen()
		assert.Equal(t, "hello", res)
		assert.True(t, q.Empty())
	}
}

func TestPushPopMany(t *testing.T) {
	q := New(10)
	for i := 0; i < 10000; i++ {
		item := fmt.Sprintf("hello%v", i)
		q.Push(item)
		res := <-q.Listen()
		assert.Equal(t, item, res)
	}
	assert.True(t, q.Empty())
}

func TestPushPopMany2(t *testing.T) {
	q := New(10)
	for i := 0; i < 10000; i++ {
		item := fmt.Sprintf("hello%v", i)
		q.Push(item)
	}
	for i := 0; i < 10000; i++ {
		item := fmt.Sprintf("hello%v", i)
		res := <-q.Listen()
		assert.Equal(t, item, res)
	}
	assert.True(t, q.Empty())
}

func TestExpand(t *testing.T) {
	q := New(10)
	// expand to 100
	for i := 0; i < 80; i++ {
		item := fmt.Sprintf("hello%v", i)
		q.Push(item)
	}
	// head is now at 40
	for i := 0; i < 40; i++ {
		item := fmt.Sprintf("hello%v", i)
		res := <-q.Listen()
		assert.Equal(t, item, res)
	}
	// make sure tail wraps around => tail is at (80+50)%100=30
	for i := 0; i < 50; i++ {
		item := fmt.Sprintf("hello%v", i+80)
		q.Push(item)
	}
	// now pop enough to make the head wrap around => (40 + 80)%100=20
	for i := 0; i < 80; i++ {
		item := fmt.Sprintf("hello%v", i+40)
		res := <-q.Listen()
		assert.Equal(t, item, res)
	}
	// push enough to cause expansion
	for i := 0; i < 100; i++ {
		item := fmt.Sprintf("hello%v", i+130)
		q.Push(item)
	}
	// empty the queue
	for i := 0; i < 110; i++ {
		item := fmt.Sprintf("hello%v", i+120)
		res := <-q.Listen()
		assert.Equal(t, item, res)
	}
	assert.True(t, q.Empty())
}

func TestQueueConsistency(t *testing.T) {
	max := 1000000
	c := 100
	var wg sync.WaitGroup
	wg.Add(1)
	q := New(10)
	go func() {
		i := 0
		seen := make(map[string]string)
		for r := range q.Listen() {
			i++
			s := r.(string)
			_, present := seen[s]
			if present {
				log.Printf("item have already been seen %v", s)
				t.FailNow()
			}
			seen[s] = s
			if i == max {
				wg.Done()
				return
			}
		}
	}()

	for j := 0; j < c; j++ {
		cmax := max / c
		jj := j
		go func() {
			for i := 0; i < cmax; i++ {
				if rand.Intn(10) == 0 {
					time.Sleep(time.Duration(rand.Intn(1000)))
				}
				q.Push(fmt.Sprintf("%v %v", jj, i))
			}
		}()
	}

	wg.Wait()
	time.Sleep(500 * time.Millisecond)
	// queue should be empty
	for i := 0; i < 100; i++ {
		select {
		case r := <-q.Listen():
			log.Printf("unexpected result %+v", r)
			t.FailNow()
		default:
		}
	}
}
