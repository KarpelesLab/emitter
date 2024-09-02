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
