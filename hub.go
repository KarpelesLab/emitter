package emitter

import (
	"context"
	"sync"
	"time"
)

var Global = &Hub{}

type Hub struct {
	Cap      uint
	topics   map[string]*topic
	topicsLk sync.RWMutex
	trig     map[string]Trigger
	trigLk   sync.RWMutex
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

func (h *Hub) getTrigger(trigName string, create bool) Trigger {
	h.trigLk.RLock()
	var t Trigger
	var ok bool
	if h.trig != nil {
		t, ok = h.trig[trigName]
	}
	h.trigLk.RUnlock()
	if ok {
		return t
	} else if !create {
		return nil
	}

	h.trigLk.Lock()
	defer h.trigLk.Unlock()

	if h.trig == nil {
		h.trig = make(map[string]Trigger)
	}

	t = NewTrigger()
	h.trig[trigName] = t
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

func (h *Hub) Push(trigger string) {
	t := h.getTrigger(trigger, false)
	if t != nil {
		t.Push()
	}
}

// Trigger returns the given trigger, creating it if needed. This can make it easy to call methods like Listen()
// or Push() in one go.
func (h *Hub) Trigger(trigName string) Trigger {
	return h.getTrigger(trigName, true)
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

// Close will turn off all of the hub's topics, ending all listeners.
func (h *Hub) Close() error {
	h.topicsLk.Lock()
	topics := h.topics
	h.topics = nil
	h.topicsLk.Unlock()
	h.trigLk.Lock()
	trig := h.trig
	h.trig = nil
	h.trigLk.Unlock()

	for _, t := range topics {
		t.close()
	}
	for _, t := range trig {
		t.Close()
	}
	return nil
}

// Emit emits an event on the given topic, and will not return until the event has been
// added to all the queues, or the context expires.
func (h *Hub) Emit(ctx context.Context, topic string, args ...any) error {
	t := h.getTopic(topic, false)
	if t == nil {
		return ErrNoSuchTopic
	}

	ev := &Event{
		Context: ctx,
		Topic:   topic,
		Args:    args,
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

// EmitEvent emits an existing [Event] object without copying it.
func (h *Hub) EmitEvent(ctx context.Context, topic string, ev *Event) error {
	ev.Topic = topic

	t := h.getTopic(topic, false)
	if t == nil {
		return ErrNoSuchTopic
	}

	return t.emit(ctx, ev)
}

// EmitEventTimeout is similar to EmitEvent but with a timeout instead of a context
func (h *Hub) EmitEventTimeout(timeout time.Duration, topic string, ev *Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return h.EmitEvent(ctx, topic, ev)
}
