package emitter

import "github.com/KarpelesLab/typutil"

type Event struct {
	Topic string
	Args  []any
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
