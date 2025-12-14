package emitter_test

import (
	"sync"
	"testing"
	"time"

	"github.com/KarpelesLab/emitter"
)

func TestNewTrigger(t *testing.T) {
	trig := emitter.NewTrigger()
	if trig == nil {
		t.Fatal("NewTrigger returned nil")
	}

	// Clean up
	trig.Close()
}

func TestTriggerPush(t *testing.T) {
	trig := emitter.NewTrigger()
	defer trig.Close()

	listener := trig.Listen()
	defer listener.Release()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		select {
		case <-listener.C:
			// success
		case <-time.After(1 * time.Second):
			t.Error("timeout waiting for trigger")
		}
	}()

	trig.Push()
	wg.Wait()
}

func TestTriggerMultiplePush(t *testing.T) {
	trig := emitter.NewTrigger()
	defer trig.Close()

	listener := trig.Listen()
	defer listener.Release()

	// Push multiple times
	for i := 0; i < 5; i++ {
		trig.Push()
	}

	// Should receive at least one (default cap is 1)
	select {
	case <-listener.C:
		// success
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for trigger")
	}
}

func TestTriggerMultipleListeners(t *testing.T) {
	trig := emitter.NewTrigger()
	defer trig.Close()

	l1 := trig.Listen()
	l2 := trig.Listen()
	l3 := trig.Listen()
	defer l1.Release()
	defer l2.Release()
	defer l3.Release()

	var wg sync.WaitGroup
	wg.Add(3)

	for _, l := range []*emitter.TriggerListener{l1, l2, l3} {
		go func(listener *emitter.TriggerListener) {
			defer wg.Done()
			select {
			case <-listener.C:
				// success
			case <-time.After(1 * time.Second):
				t.Error("timeout waiting for trigger")
			}
		}(l)
	}

	trig.Push()
	wg.Wait()
}

func TestTriggerListenCap(t *testing.T) {
	trig := emitter.NewTrigger()
	defer trig.Close()

	// Create listener with capacity 5
	listener := trig.ListenCap(5)
	defer listener.Release()

	// Push 5 times - all should queue
	for i := 0; i < 5; i++ {
		trig.Push()
	}

	// Give time for pushes to be processed
	time.Sleep(50 * time.Millisecond)

	// Should receive all 5
	received := 0
	for {
		select {
		case <-listener.C:
			received++
			if received == 5 {
				return
			}
		case <-time.After(100 * time.Millisecond):
			if received < 5 {
				t.Errorf("only received %d triggers, expected 5", received)
			}
			return
		}
	}
}

func TestTriggerRelease(t *testing.T) {
	trig := emitter.NewTrigger()
	defer trig.Close()

	listener := trig.Listen()

	// Release should not panic
	listener.Release()

	// Push after release should not block or panic
	trig.Push()
}

func TestTriggerClose(t *testing.T) {
	trig := emitter.NewTrigger()

	listener := trig.Listen()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for range listener.C {
			// drain until closed
		}
	}()

	// Close trigger
	if err := trig.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Listener channel should close
	wg.Wait()
}

func TestTriggerCloseMultipleListeners(t *testing.T) {
	trig := emitter.NewTrigger()

	l1 := trig.Listen()
	l2 := trig.Listen()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for range l1.C {
		}
	}()

	go func() {
		defer wg.Done()
		for range l2.C {
		}
	}()

	trig.Close()
	wg.Wait()
}

func TestTriggerPushNoListeners(t *testing.T) {
	trig := emitter.NewTrigger()
	defer trig.Close()

	// Push without listeners should not panic or block
	trig.Push()
}

func TestTriggerConcurrentPush(t *testing.T) {
	trig := emitter.NewTrigger()
	defer trig.Close()

	listener := trig.ListenCap(100)
	defer listener.Release()

	var wg sync.WaitGroup

	// Concurrent pushes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				trig.Push()
			}
		}()
	}

	wg.Wait()

	// Drain any queued triggers
	drained := 0
	for {
		select {
		case <-listener.C:
			drained++
		case <-time.After(100 * time.Millisecond):
			// Done draining
			if drained == 0 {
				t.Error("expected at least some triggers")
			}
			return
		}
	}
}

func TestTriggerListenerChannel(t *testing.T) {
	trig := emitter.NewTrigger()
	defer trig.Close()

	listener := trig.Listen()
	defer listener.Release()

	if listener.C == nil {
		t.Error("listener channel is nil")
	}
}
