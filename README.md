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

# Trigger

The trigger object allows waking multiple threads at the same time using channels rather than [sync.Cond](https://pkg.go.dev/sync#Cond). This can be useful to wake many threads to specific events while still using other event sources such as timers.

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
        case <-l.C:
            // do something on trigger called
        }
    }
}

trig.Push() // push event
```
