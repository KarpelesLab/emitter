package emitter

type listener struct {
	ch chan *Event
}

func newListener(c uint) *listener {
	res := &listener{
		ch: make(chan *Event, c),
	}
	return res
}
