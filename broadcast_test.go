package broadcast

import (
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var listenCount int32

func listen(r Receiver) {
	for v := r.Read(); v != nil; v = r.Read() {
		atomic.AddInt32(&listenCount, 1)
		go listen(r)
	}
}

func TestSingleListener(t *testing.T) {
	b := NewBroadcaster()
	r := b.Listen()

	go listen(r)

	const max = 10
	// count the expect listenCount
	var expect int32
	for i := 0; i < max; i++ {
		expect += int32(math.Pow(2, float64(i)))
	}

	for i := 0; i < max; i++ {
		b.Write(i)
	}

	// sleep enough time to wait all done
	time.Sleep(1 * time.Second)

	// check the result
	result := atomic.LoadInt32(&listenCount)
	if result != expect {
		t.Errorf("expect %d, got %d\n", expect, result)
	}

	b.Close()
}

func listener(r Receiver, done func(), result chan<- []interface{}) {
	defer done()
	stuff := make([]interface{}, 0)
	for v := r.Read(); v != nil; v = r.Read() {
		stuff = append(stuff, v)
	}
	result <- stuff
}

func compareSlice(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i, ae := range a {
		if ae != b[i] {
			return false
		}
	}
	return true
}

func TestMultiListeners(t *testing.T) {
	b := NewBroadcaster()
	const listeners = 100
	receives := make([]chan []interface{}, listeners)
	waiter := new(sync.WaitGroup)
	done := func() {
		waiter.Done()
	}
	for i := 0; i < listeners; i++ {
		receives[i] = make(chan []interface{}, 1)
		waiter.Add(1)
		go listener(b.Listen(), done, receives[i])
	}

	samples := []interface{}{1, 2}
	for _, sample := range samples {
		b.Write(sample)
	}

	b.Close()
	waiter.Wait()
	for i := 0; i < listeners; i++ {
		result := <-receives[i]
		if compareSlice(result, samples) == false {
			t.Errorf("%d listerner: expect %v, got %v\n", i, samples, result)
		}
	}
}
