package broadcast

import (
	"fmt"
	"sync"
	"testing"
)

func listen(r *Receiver, done func(), wait *sync.WaitGroup) {
	defer done()
	for v := r.Read(); v != nil; v = r.Read() {
		wait.Add(1)
		fmt.Println(v)
		go listen(r, done, wait)
	}
}

func TestBroadcast(t *testing.T) {
	b := NewBroadcaster()
	r := b.Listen()
	listernerCount := 0

	wait := new(sync.WaitGroup)
	wait.Add(1)
	done := func() {
		listernerCount++
		println(listernerCount)
		wait.Done()
	}
	go listen(r, done, wait)

	for i := 0; i < 2; i++ {
		b.Write(i)
	}
	b.Write(nil)
	wait.Wait()
}
