[![GoDoc](https://godoc.org/github.com/KarpelesLab/emitter?status.svg)](https://godoc.org/github.com/KarpelesLab/emitter)

# emitter

Simple emitter lib.

## Example

```go
h := emitter.New()

go func(ch <-chan *emitter.Event) {
    for ev := range ch {
        // handle event
        intVal, err := emitter.Arg[int](ev, 0)
        // intVal is 42
    }
}(h.On("event"))

h.Emit("event", 42)
```

## Global Hub

For some cases it might be useful to have global events that can be handled by multiple packages. This can be done with the global hub.

```go
go func(ch <-chan *emitter.Event) {
    for ev := range ch {
        // ...
    }
})(emitter.Global.On("event"))

emitter.Global.Emit("event", 42)
```

# Trigger

The trigger object allows waking multiple threads at the same time using channels rather than [sync.Cond](https://pkg.go.dev/sync#Cond). This can be useful to wake many threads to specific events while still using other event sources such as timers.

Triggers by default have a queue size of 1, meaning that a call to Push() can be queued and delivered later if the receiving thread is busy. Cap can be set to other values including zero (do not queue) or larger values (queue a number of calls).

## Example

```go
trig := emitter.NewTrigger()

go func() {
    t := time.NewTicker(30*time.Second)
    defer t.Stop()
    l := trig.Listen()
    defer l.Release()

    for {
        select {
        case <-t.C:
            // do something every 30 secs
        case _, closed := <-l.C:
            if closed {
                // exit loop if channel is closed, would be triggered by trig.Close()
                // this can be ignored if trig.Closed() isn't called
                return
            }
            // do something on trigger called
        }
    }
}

trig.Push() // push event
```
