package emitter

import (
	"context"
	"errors"
	"sync"

	"github.com/KarpelesLab/typutil"
)

// ArgEncoder is a function type for encoding event arguments into bytes.
// It is used with [Event.EncodedArg] to cache encoded representations of
// arguments, such as JSON encoding.
type ArgEncoder func(any) ([]byte, error)

// Event represents an emitted event that is delivered to subscribers.
// It contains the event context, topic name, and any arguments passed
// during emission.
type Event struct {
	// Context is the context passed to [Hub.Emit] or [Hub.EmitEvent].
	// It can be used for cancellation or deadline propagation.
	Context context.Context

	// Topic is the name of the topic this event was emitted on.
	Topic string

	// Args contains the arguments passed to [Hub.Emit].
	Args []any

	argAs   []map[string]*encodedArg
	argAsLk sync.Mutex
}

type encodedArg struct {
	once sync.Once
	buf  []byte
	err  error
}

// Arg returns the nth argument from the event, or nil if n is out of bounds.
// For type-safe argument access with conversion, use the generic [Arg] function.
func (ev *Event) Arg(n uint) any {
	if n >= uint(len(ev.Args)) {
		return nil
	}
	return ev.Args[n]
}

// Arg is a generic function that returns the nth argument from an event,
// converted to the specified type T. It uses type conversion and will
// return an error if the argument cannot be converted to type T.
//
// Example:
//
//	intVal, err := emitter.Arg[int](ev, 0)
//	strVal, err := emitter.Arg[string](ev, 1)
func Arg[T any](ev *Event, arg uint) (T, error) {
	return typutil.As[T](ev.Arg(arg))
}

func (ev *Event) getEncodedArg(n uint, key string) *encodedArg {
	ev.argAsLk.Lock()
	defer ev.argAsLk.Unlock()

	if n >= uint(len(ev.Args)) {
		return nil
	}

	if ev.argAs == nil {
		ev.argAs = make([]map[string]*encodedArg, len(ev.Args))
	}

	if ev.argAs[n] == nil {
		list := make(map[string]*encodedArg)
		ev.argAs[n] = list
		obj := &encodedArg{}
		list[key] = obj
		return obj
	}

	if obj, ok := ev.argAs[n][key]; ok {
		return obj
	}
	obj := &encodedArg{}
	ev.argAs[n][key] = obj
	return obj
}

// EncodedArg returns an encoded representation of the nth argument, using the
// provided encoder function. The result is cached by key, so subsequent calls
// with the same argument index and key will return the cached result without
// re-encoding. This is useful for efficiently encoding arguments to formats
// like JSON when multiple listeners need the same encoded representation.
//
// Example:
//
//	jsonBytes, err := ev.EncodedArg(0, "json", json.Marshal)
func (ev *Event) EncodedArg(n uint, key string, encoder ArgEncoder) ([]byte, error) {
	ea := ev.getEncodedArg(n, key)
	if ea == nil {
		return nil, errors.New("invalid event argument number")
	}
	ea.once.Do(func() {
		ea.buf, ea.err = encoder(ev.Arg(n))
	})
	return ea.buf, ea.err
}
