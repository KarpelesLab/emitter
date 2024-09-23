package emitter

import (
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
)

// Trigger is a simple lightweight process for sending simple one shot notifications on
// multiple channels. There is no event or notification content, just a "wake up" signal
// sent to many channels at the same time. Calling the trigger costs virtually nothing
// while listeners can have more complex logic.
//
// Even if there are a lot of listeners and it takes time to deliver notifications, each
// call to Push will translate to one attempt to push something on the listening channels.
// The only case where a channel drops notifications is if it's not ready to listen, but
// this can be solved by adding some capacity to the channels by setting Cap to something
// larger than zero.
type Trigger struct {
	Cap    uint   // capacity for channels generated by Trigger.Listen
	send   uint32 // number of pending signals
	l      sync.RWMutex
	c      *sync.Cond
	closed uint32 // if not zero, means this trigger has been closed
	ch     map[<-chan struct{}]chan struct{}
	chLk   sync.RWMutex
}

// TriggerListener represents a listener that will receive notifications when the trigger
// is pushed. Call Release() after using it to close the channel (with a defer l.Release())
type TriggerListener struct {
	C <-chan struct{}
	t *Trigger
}

// NewTrigger returns a new trigger object ready for use. This will also create a goroutine
func NewTrigger() *Trigger {
	tr := &Trigger{
		Cap: 1, // 1 by default so we can queue even just 1 pending call
		ch:  make(map[<-chan struct{}]chan struct{}),
	}
	tr.c = sync.NewCond(tr.l.RLocker())
	go tr.thread()
	return tr
}

// Push will wake all the listeners for this trigger
func (t *Trigger) Push() {
	atomic.AddUint32(&t.send, 1)
	t.c.Broadcast()
}

// Close will close all the listeners for this trigger
func (t *Trigger) Close() error {
	atomic.AddUint32(&t.closed, 1)
	t.Push()
	return nil
}

// Listen returns a listener object. Remember to release the object after you stop using it
// (calling release will close the channel).
func (t *Trigger) Listen() *TriggerListener {
	c := make(chan struct{}, t.Cap)
	res := &TriggerListener{
		C: c,
		t: t,
	}

	runtime.SetFinalizer(res, releaseTriggerListener)

	t.chLk.Lock()
	defer t.chLk.Unlock()
	t.ch[c] = c
	return res
}

func releaseTriggerListener(o *TriggerListener) {
	o.Release()
}

// Release will close the channel linked to this trigger and stop sending it messages
func (tl *TriggerListener) Release() {
	t := tl.t
	t.chLk.Lock()
	defer t.chLk.Unlock()
	if c, ok := t.ch[tl.C]; ok {
		delete(t.ch, tl.C)
		close(c)
	}
}

var emptyStructVal = reflect.ValueOf(struct{}{})

// makeCases generates an array of [reflect.SelectCase] for use with reflect.Select
func (t *Trigger) makeCases() []reflect.SelectCase {
	t.chLk.RLock()
	defer t.chLk.RUnlock()

	if len(t.ch) == 0 {
		// do nothing
		return nil
	}

	res := make([]reflect.SelectCase, len(t.ch)+1)
	res[0].Dir = reflect.SelectDefault

	n := 1

	for _, l := range t.ch {
		res[n].Dir = reflect.SelectSend
		res[n].Chan = reflect.ValueOf(l)
		res[n].Send = emptyStructVal
		n += 1
	}

	return res
}

// emit pushes a struct{}{} on all known channels
func (t *Trigger) emit() {
	cases := t.makeCases()
	if len(cases) == 0 {
		// do nothing
		return
	}
	cnt := len(cases) - 1 // number of sends we expect, considering cases[0] is reserved for context timeout

	for {
		// (chosen int, recv Value, recvOK bool)
		chosen, _, _ := reflect.Select(cases)
		if chosen == 0 {
			// default, meaning there is nothing ready to accept anymore
			return
		}
		cnt -= 1
		if cnt == 0 {
			// all sends completed successfully
			return
		}
		// set to nil & continue
		cases[chosen].Chan = reflect.Value{}
	}
}

// thread is a thread just running and that's all
func (t *Trigger) thread() {
	t.l.RLock()
	defer t.l.RUnlock()

	for {
		if atomic.LoadUint32(&t.closed) != 0 {
			// closing!
			t.chLk.Lock()
			defer t.chLk.Unlock()
			for _, c := range t.ch {
				close(c)
			}
			clear(t.ch)
			return
		}
		if atomic.LoadUint32(&t.send) == 0 {
			// only wait if there is nothing to send
			t.c.Wait()
			continue
		}
		// at this point t.send is guaranteed to be >0
		t.emit()
		atomic.AddUint32(&t.send, ^uint32(0))
	}
}
