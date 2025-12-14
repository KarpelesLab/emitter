// Package emitter provides a lightweight, thread-safe event emission library for Go
// that implements a publish-subscribe pattern. It offers two main mechanisms for
// inter-goroutine communication:
//
//   - Event System: For sending rich events with data and context
//   - Trigger System: For simple, fast one-shot notifications
//
// # Event System
//
// The event system allows publishing events to named topics. Subscribers receive
// events through channels that they can select on alongside other channel operations.
//
// Basic usage:
//
//	h := emitter.New()
//	ch := h.On("event")
//
//	go func() {
//	    defer h.Off("event", ch)
//	    for ev := range ch {
//	        intVal, err := emitter.Arg[int](ev, 0)
//	        // process event...
//	    }
//	}()
//
//	h.Emit(context.Background(), "event", 42)
//
// A global [Hub] is available via [Global] for cases where events need to be
// shared across multiple packages.
//
// # Trigger System
//
// Triggers provide a lightweight mechanism to wake multiple goroutines without
// carrying event data. Unlike events, triggers use atomic operations for minimal
// overhead and return immediately without blocking.
//
// Trigger usage:
//
//	trig := emitter.NewTrigger()
//
//	go func() {
//	    l := trig.Listen()
//	    defer l.Release()
//	    for {
//	        select {
//	        case <-l.C:
//	            // handle trigger
//	        }
//	    }
//	}()
//
//	trig.Push() // wake all listeners
//
// Triggers support configurable channel capacity to allow queuing of notifications.
// The default capacity is 1, allowing one pending notification to be queued.
//
// # Thread Safety
//
// All types in this package are safe for concurrent use by multiple goroutines.
package emitter
