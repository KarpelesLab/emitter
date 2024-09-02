package emitter

import (
	"context"
	"reflect"
	"sync"
)

type topic struct {
	listeners   map[<-chan *Event]*listener
	listenersLk sync.RWMutex
}

func newTopic() *topic {
	res := &topic{
		listeners: make(map[<-chan *Event]*listener),
	}
	return res
}

func (t *topic) appendListener(l *listener) {
	t.listenersLk.Lock()
	defer t.listenersLk.Unlock()
	t.listeners[l.ch] = l
}

func (t *topic) newListener(c uint) <-chan *Event {
	l := newListener(c)
	t.appendListener(l)
	return l.ch
}

func (t *topic) takeAll() []*listener {
	t.listenersLk.Lock()
	defer t.listenersLk.Unlock()

	res := make([]*listener, 0, len(t.listeners))
	for _, l := range t.listeners {
		res = append(res, l)
	}

	clear(t.listeners)
	return res
}

func (t *topic) makeCases(ctx context.Context, val reflect.Value) []reflect.SelectCase {
	t.listenersLk.RLock()
	defer t.listenersLk.RUnlock()

	res := make([]reflect.SelectCase, len(t.listeners)+1)
	res[0].Dir = reflect.SelectRecv

	if ch := ctx.Done(); ch != nil {
		res[0].Chan = reflect.ValueOf(ch)
	}

	n := 1

	for _, l := range t.listeners {
		res[n].Dir = reflect.SelectSend
		res[n].Chan = reflect.ValueOf(l.ch)
		res[n].Send = val
		n += 1
	}

	return res
}

func (t *topic) emit(ctx context.Context, ev *Event) error {
	cases := t.makeCases(ctx, reflect.ValueOf(ev))
	cnt := len(cases) - 1 // number of sends we expect, considering cases[0] is reserved for context timeout

	for {
		// (chosen int, recv Value, recvOK bool)
		chosen, _, _ := reflect.Select(cases)
		if chosen == 0 {
			// ctx.Done()
			return ctx.Err()
		}
		cnt -= 1
		if cnt == 0 {
			// all sends completed successfully
			return nil
		}
		// set to nil & continue
		cases[chosen].Chan = reflect.Value{}
	}
}

func (t *topic) close() {
	ls := t.takeAll()
	for _, l := range ls {
		close(l.ch)
	}
}

func (t *topic) remove(ch <-chan *Event) {
	t.listenersLk.Lock()
	defer t.listenersLk.Unlock()

	if l, ok := t.listeners[ch]; ok {
		delete(t.listeners, ch)
		close(l.ch)
	}
}
