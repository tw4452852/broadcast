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
}

func NewBroadcaster() *Broadcaster {
	listenC := make(chan chan chan broadcast)
	sendC := make(chan interface{})

	go func() {
		currc := make(chan broadcast, 1)
		for {
			select {
			case v := <-sendC:
				if v == nil {
					currc <- broadcast{}
					return
				}
				c := make(chan broadcast, 1)
				b := broadcast{c: c, v: v}
				currc <- b
				currc = c
			case r := <-listenC:
				r <- currc
			}
		}
	}()

	return &Broadcaster{
		listenC: listenC,
		sendC:   sendC,
	}
}

// Write send a broadcast message to all listeners
func (b *Broadcaster) Write(v interface{}) {
	b.sendC <- v
}

// Listen register a listerner
func (b *Broadcaster) Listen() *Receiver {
	c := make(chan chan broadcast)
	b.listenC <- c
	return &Receiver{<-c}
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
