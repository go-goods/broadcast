package broadcast

//Sender is a type that can have messages Sent to it.
type Sender interface {
	Send(interface{})
}

//Receiver is a type that can receive a message and an optional ok specifying
//if the Receiver is closed, as in the case of a channel receive.
type Receiver interface {
	Receive() (interface{}, bool)
}

//Broadcaster represents a type that takes values from a Receiver and sends them
//to many Senders.
type Broadcaster struct {
	msgs       chan interface{}
	register   chan Sender
	unregister chan Sender
	quit       chan bool
}

func (b *Broadcaster) pipe(input Receiver) {
	for {
		in, ok := input.Receive()
		if !ok {
			close(b.quit)
			return
		}
		b.msgs <- in
	}
}

func (b *Broadcaster) run() {
	ls := make(map[Sender]struct{})
	for {
		select {
		case s := <-b.register:
			ls[s] = struct{}{}
		case s := <-b.unregister:
			delete(ls, s)
		case m := <-b.msgs:
			for s := range ls {
				s.Send(m)
			}
		case <-b.quit:
			return
		}
	}
}

//Register takes a Sender and will call the Send method on it every time a
//message is received from the Receiver.
func (b *Broadcaster) Register(s Sender) {
	b.register <- s
}

//Unregister removes the Sender from the set that gets the Send method called
//for each message received from the Receiver.
func (b *Broadcaster) Unregister(s Sender) {
	b.unregister <- s
}

//New creates a new Broadcaster that calls the Receive method on the given
//Receiver and forwards that message to each registered Sender. If the Receiver
//ever returns false as the second parameter, the Broadcaster shuts down and is
//cleaned up. The Send calls are synchronous, and no new Register or Unregister
//call will proceed until every Send call finishes.
func New(r Receiver) (b *Broadcaster) {
	b = &Broadcaster{
		msgs:       make(chan interface{}),
		register:   make(chan Sender),
		unregister: make(chan Sender),
		quit:       make(chan bool),
	}
	go b.pipe(r)
	go b.run()
	return
}
