package broadcast

// broadcast convey a message
type broadcast struct {
	c chan broadcast
	v interface{}
}

// a broadcast center
type Broadcaster struct {
	listenC chan chan chan broadcast
	sendC   chan<- interface{}
	closeC  chan bool
}

func NewBroadcaster() *Broadcaster {
	listenC := make(chan chan chan broadcast)
	sendC := make(chan interface{})
	closeC := make(chan bool)

	go func() {
		currc := make(chan broadcast, 1)
		isclose := false
		for {
			select {
			case v := <-sendC:
				if isclose {
					// nothing to do
					break
				}
				c := make(chan broadcast, 1)
				b := broadcast{c: c, v: v}
				currc <- b
				currc = c
			case r := <-listenC:
				r <- currc
			case <-closeC:
				if isclose == false {
					// inform the listeners we are closing
					currc <- broadcast{}
				}
				isclose = true
			}
		}
	}()

	return &Broadcaster{
		listenC: listenC,
		sendC:   sendC,
		closeC:  closeC,
	}
}

// Write send a broadcast message to all listeners
func (b *Broadcaster) Write(v interface{}) {
	b.sendC <- v
}

// Listen register a listerner
func (b *Broadcaster) Listen() Receiver {
	c := make(chan chan broadcast)
	b.listenC <- c
	return Receiver{<-c}
}

// close the broadcaster
func (b *Broadcaster) Close() {
	b.closeC <- true
}

type Receiver struct {
	c chan broadcast
}

// Read a value that has been broadcast,
// wait until one is available if necessary
func (r *Receiver) Read() interface{} {
	b := <-r.c
	v := b.v
	r.c <- b
	r.c = b.c
	return v
}
