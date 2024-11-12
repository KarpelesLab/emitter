package emitter

import "sync"

type listener struct {
	ch chan *Event
	lk sync.Mutex
}

func newListener(c uint) *listener {
	res := &listener{
		ch: make(chan *Event, c),
	}
	return res
}

func (l *listener) close() {
	l.lk.Lock()
	defer l.lk.Unlock()

	if l.ch != nil {
		close(l.ch)
		l.ch = nil
	}
}
