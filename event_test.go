package emitter_test

import (
	"encoding/json"
	"testing"

	"github.com/KarpelesLab/emitter"
)

func TestEventArg(t *testing.T) {
	ev := &emitter.Event{Args: []any{"42", "hello"}}

	v, err := emitter.Arg[int](ev, 0)
	if err != nil {
		t.Errorf("failed conversion: %s", err)
	}
	if v != 42 {
		t.Errorf("unexpected value %d", v)
	}

	strV, err := emitter.Arg[string](ev, 1)
	if err != nil {
		t.Errorf("failed conversion: %s", err)
	}
	if strV != "hello" {
		t.Errorf("unexpected value %v", strV)
	}

	strJ, err := ev.EncodedArg(1, "json", json.Marshal)
	if err != nil {
		t.Errorf("failed json encode: %s", err)
	}
	if string(strJ) != `"hello"` {
		t.Errorf("invalid json: %s", strJ)
	}
}
