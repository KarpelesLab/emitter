package emitter

import (
	"context"
	"errors"
	"sync"

	"github.com/KarpelesLab/typutil"
)

type ArgEncoder func(any) ([]byte, error)

type Event struct {
	Context context.Context
	Topic   string
	Args    []any
	argAs   []map[string]*encodedArg
	argAsLk sync.Mutex
}

type encodedArg struct {
	once sync.Once
	buf  []byte
	err  error
}

func (ev *Event) Arg(n uint) any {
	if n >= uint(len(ev.Args)) {
		return nil
	}
	return ev.Args[n]
}

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
