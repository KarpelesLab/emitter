[![GoDoc](https://godoc.org/github.com/KarpelesLab/emitter?status.svg)](https://godoc.org/github.com/KarpelesLab/emitter)

# emitter

A lightweight, thread-safe event emission library for Go that implements a publish-subscribe pattern.

## Features

- **Event System**: Send rich events with data and context to named topics
- **Trigger System**: Fast, lightweight one-shot notifications using atomic operations
- **Thread-Safe**: All operations are safe for concurrent use
- **Context Support**: Events respect context deadlines and cancellation
- **Type-Safe Arguments**: Generic `Arg[T]()` function for type-safe argument access
- **Flexible Buffering**: Configurable channel capacity for both events and triggers

## Installation

```bash
go get github.com/KarpelesLab/emitter
```

## Event System

### Basic Example

```go
h := emitter.New()
ch := h.On("event")

go func() {
    defer h.Off("event", ch)

    for ev := range ch {
        // handle event
        intVal, err := emitter.Arg[int](ev, 0)
        // intVal is 42
    }
}()

h.Emit(context.Background(), "event", 42)
```

### Global Hub

For cases where events need to be shared across multiple packages, use the global hub:

```go
ch := emitter.Global.On("event")

go func() {
    for ev := range ch {
        // ...
    }
}()

emitter.Global.Emit(context.Background(), "event", 42)
```

### Emitting with Timeout

For goroutine-based emission with a timeout instead of context:

```go
go h.EmitTimeout(30*time.Second, "topic", args...)
```

## Trigger System

The trigger object allows waking multiple goroutines at the same time using channels rather than [sync.Cond](https://pkg.go.dev/sync#Cond). This is useful for waking many goroutines to specific events while still using other event sources such as timers.

Triggers by default have a queue size of 1, meaning that a call to `Push()` can be queued and delivered later if the receiving goroutine is busy. Capacity can be set to other values including zero (do not queue) or larger values (queue multiple calls).

Unlike events `Emit()`, a trigger's `Push()` method returns instantly and is non-blocking, with minimal resource usage (a single atomic operation).

### Example

```go
trig := emitter.Global.Trigger("test")

go func() {
    t := time.NewTicker(30*time.Second)
    defer t.Stop()
    l := trig.Listen()
    defer l.Release()

    for {
        select {
        case <-t.C:
            // do something every 30 secs
        case <-l.C:
            // do something on trigger called
        }
    }
}()

trig.Push() // push signal to all listeners
```

### Standalone Triggers

Triggers can also be used independently of a Hub:

```go
trig := emitter.NewTrigger()

l := trig.Listen()
defer l.Release()

// In another goroutine
trig.Push()
```

## API Overview

### Hub Methods

| Method | Description |
|--------|-------------|
| `New()` | Create a new Hub instance |
| `On(topic)` | Subscribe to a topic, returns a channel |
| `OnWithCap(topic, cap)` | Subscribe with custom channel capacity |
| `Off(topic, ch)` | Unsubscribe from a topic |
| `Emit(ctx, topic, args...)` | Emit an event (blocks until delivered or context expires) |
| `EmitTimeout(timeout, topic, args...)` | Emit with timeout |
| `Trigger(name)` | Get or create a named trigger |
| `Push(name)` | Push signal to a named trigger |
| `Close()` | Close all topics and triggers |

### Event Methods

| Method | Description |
|--------|-------------|
| `Arg(n)` | Get nth argument as `any` |
| `Arg[T](ev, n)` | Get nth argument with type conversion |
| `EncodedArg(n, key, encoder)` | Get cached encoded representation |

### Trigger Interface

| Method | Description |
|--------|-------------|
| `Listen()` | Create a listener with default capacity |
| `ListenCap(cap)` | Create a listener with custom capacity |
| `Push()` | Wake all listeners (non-blocking) |
| `Close()` | Close trigger and all listeners |

## License

MIT License - see [LICENSE](LICENSE) file.
