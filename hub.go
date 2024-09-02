package emitter

import (
	"context"
	"sync"
	"time"
)

type Hub struct {
	Cap      uint
	topics   map[string]*topic
	topicsLk sync.RWMutex
}

func New() *Hub {
	return &Hub{}
}

func (h *Hub) getTopic(topicName string, create bool) *topic {
	h.topicsLk.RLock()
	var t *topic
	var ok bool
	if h.topics != nil {
		t, ok = h.topics[topicName]
	}
	h.topicsLk.RUnlock()
	if ok {
		return t
	} else if !create {
		return nil
	}

	h.topicsLk.Lock()
	defer h.topicsLk.Unlock()

	if h.topics == nil {
		h.topics = make(map[string]*topic)
	}

	t = newTopic()
	h.topics[topicName] = t
	return t
}

// On returns a channel that will receive events
func (h *Hub) On(topic string) <-chan *Event {
	return h.getTopic(topic, true).newListener(h.Cap)
}

// OnWithCap returns a channel that will receive events, and has the given capacity instead of the default one
func (h *Hub) OnWithCap(topic string, c uint) <-chan *Event {
	return h.getTopic(topic, true).newListener(c)
}

// Off unsubscribes from a given topic. If ch is nil, the whole topic is closed, otherwise only the given
// channel is removed from the topic. Note that the channel will be closed in the process.
func (h *Hub) Off(topic string, ch <-chan *Event) {
	t := h.getTopic(topic, false)
	if t == nil {
		return
	}

	if ch == nil {
		// close whole topic
		t.close()
		return
	} else {
		t.remove(ch)
	}
}

// Emit emits an event on the given topic, and will not return until the event has been
// added to all the queues, or the context expires.
func (h *Hub) Emit(ctx context.Context, topic string, args ...any) error {
	t := h.getTopic(topic, false)
	if t == nil {
		return ErrNoSuchTopic
	}

	ev := &Event{
		Topic: topic,
		Args:  args,
	}

	return t.emit(ctx, ev)
}

// EmitTimeout emits an event with a given timeout instead of using a context. This is useful
// when emitting events in goroutines, ie:
//
//	go h.EmitTimeout(30*time.Second, "topic", args...)
func (h *Hub) EmitTimeout(timeout time.Duration, topic string, args ...any) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return h.Emit(ctx, topic, args...)
}
