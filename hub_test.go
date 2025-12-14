package emitter_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/KarpelesLab/emitter"
)

func TestEvents(t *testing.T) {
	h := emitter.New()

	ch := h.On("test1")
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		ev := <-ch
		if ev.Topic != "test1" {
			t.Errorf("unexpected topic: %s", ev.Topic)
		}
		arg, err := emitter.Arg[string](ev, 0)
		if err != nil {
			t.Errorf("failed to get arg: %v", err)
		}
		if arg != "hello world" {
			t.Errorf("unexpected arg: %s", arg)
		}
	}()

	if err := h.Emit(context.Background(), "test1", "hello world"); err != nil {
		t.Errorf("Emit failed: %v", err)
	}

	wg.Wait()
}

func TestOnWithCap(t *testing.T) {
	h := emitter.New()

	ch := h.OnWithCap("test", 5)
	if ch == nil {
		t.Fatal("OnWithCap returned nil channel")
	}

	// Emit multiple events - with capacity 5, they should all queue
	for i := 0; i < 5; i++ {
		go func(i int) {
			_ = h.Emit(context.Background(), "test", i)
		}(i)
	}

	// Give time for events to be queued
	time.Sleep(50 * time.Millisecond)

	// Receive all events
	received := 0
	for received < 5 {
		select {
		case <-ch:
			received++
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("timeout waiting for events, received %d", received)
		}
	}
}

func TestOff(t *testing.T) {
	h := emitter.New()

	ch := h.On("test")

	// Start a goroutine to receive
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range ch {
			// drain channel until closed
		}
	}()

	// Emit an event
	if err := h.Emit(context.Background(), "test", "data"); err != nil {
		t.Errorf("Emit failed: %v", err)
	}

	// Unsubscribe
	h.Off("test", ch)

	// Wait for channel to close
	wg.Wait()
}

func TestOffCloseTopic(t *testing.T) {
	h := emitter.New()

	ch1 := h.On("test")
	ch2 := h.On("test")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for range ch1 {
		}
	}()

	go func() {
		defer wg.Done()
		for range ch2 {
		}
	}()

	// Close entire topic by passing nil
	h.Off("test", nil)

	// Both channels should close
	wg.Wait()
}

func TestOffNonExistentTopic(t *testing.T) {
	h := emitter.New()

	// Should not panic
	h.Off("nonexistent", nil)
}

func TestClose(t *testing.T) {
	h := emitter.New()

	ch1 := h.On("topic1")
	ch2 := h.On("topic2")
	trig := h.Trigger("trigger1")
	listener := trig.Listen()

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		for range ch1 {
		}
	}()

	go func() {
		defer wg.Done()
		for range ch2 {
		}
	}()

	go func() {
		defer wg.Done()
		for range listener.C {
		}
	}()

	// Close entire hub
	if err := h.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// All channels should close
	wg.Wait()
}

func TestEmitTimeout(t *testing.T) {
	h := emitter.New()

	ch := h.On("test")

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		ev := <-ch
		arg, _ := emitter.Arg[int](ev, 0)
		if arg != 42 {
			t.Errorf("unexpected arg: %d", arg)
		}
	}()

	if err := h.EmitTimeout(1*time.Second, "test", 42); err != nil {
		t.Errorf("EmitTimeout failed: %v", err)
	}

	wg.Wait()
}

func TestEmitTimeoutExpires(t *testing.T) {
	h := emitter.New()

	// Create channel but don't read from it
	_ = h.On("test")

	// Should timeout because no one is reading
	err := h.EmitTimeout(10*time.Millisecond, "test", "data")
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestEmitEvent(t *testing.T) {
	h := emitter.New()

	ch := h.On("test")

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		ev := <-ch
		if ev.Topic != "test" {
			t.Errorf("unexpected topic: %s", ev.Topic)
		}
	}()

	ev := &emitter.Event{
		Context: context.Background(),
		Args:    []any{"data"},
	}

	if err := h.EmitEvent(context.Background(), "test", ev); err != nil {
		t.Errorf("EmitEvent failed: %v", err)
	}

	wg.Wait()
}

func TestEmitEventTimeout(t *testing.T) {
	h := emitter.New()

	ch := h.On("test")

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		<-ch
	}()

	ev := &emitter.Event{
		Args: []any{"data"},
	}

	if err := h.EmitEventTimeout(1*time.Second, "test", ev); err != nil {
		t.Errorf("EmitEventTimeout failed: %v", err)
	}

	wg.Wait()
}

func TestEmitNoSuchTopic(t *testing.T) {
	h := emitter.New()

	err := h.Emit(context.Background(), "nonexistent", "data")
	if err != emitter.ErrNoSuchTopic {
		t.Errorf("expected ErrNoSuchTopic, got: %v", err)
	}
}

func TestEmitEventNoSuchTopic(t *testing.T) {
	h := emitter.New()

	ev := &emitter.Event{Args: []any{"data"}}
	err := h.EmitEvent(context.Background(), "nonexistent", ev)
	if err != emitter.ErrNoSuchTopic {
		t.Errorf("expected ErrNoSuchTopic, got: %v", err)
	}
}

func TestHubTrigger(t *testing.T) {
	h := emitter.New()

	trig := h.Trigger("test")
	if trig == nil {
		t.Fatal("Trigger returned nil")
	}

	// Getting same trigger should return same instance
	trig2 := h.Trigger("test")
	if trig != trig2 {
		t.Error("expected same trigger instance")
	}
}

func TestHubPush(t *testing.T) {
	h := emitter.New()

	trig := h.Trigger("test")
	listener := trig.Listen()

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

	// Use Hub.Push
	h.Push("test")

	wg.Wait()
}

func TestHubPushNonExistent(t *testing.T) {
	h := emitter.New()

	// Should not panic
	h.Push("nonexistent")
}

func TestEmitContextCanceled(t *testing.T) {
	h := emitter.New()

	// Create channel but don't read from it
	_ = h.On("test")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := h.Emit(ctx, "test", "data")
	if err == nil {
		t.Error("expected context canceled error")
	}
}

func TestGlobalHub(t *testing.T) {
	// Test that Global hub works
	if emitter.Global == nil {
		t.Fatal("Global hub is nil")
	}

	ch := emitter.Global.On("global_test")

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		<-ch
	}()

	if err := emitter.Global.Emit(context.Background(), "global_test", "data"); err != nil {
		t.Errorf("Global.Emit failed: %v", err)
	}

	wg.Wait()

	// Cleanup
	emitter.Global.Off("global_test", ch)
}

func TestHubCap(t *testing.T) {
	h := emitter.New()
	h.Cap = 3

	ch := h.On("test")

	// With Cap=3, we should be able to emit 3 events without blocking
	for i := 0; i < 3; i++ {
		go func(i int) {
			_ = h.Emit(context.Background(), "test", i)
		}(i)
	}

	time.Sleep(50 * time.Millisecond)

	// Drain
	for i := 0; i < 3; i++ {
		select {
		case <-ch:
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("timeout at event %d", i)
		}
	}
}
