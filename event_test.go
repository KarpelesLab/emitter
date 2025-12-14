package emitter_test

import (
	"encoding/json"
	"errors"
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

func TestEventArgOutOfBounds(t *testing.T) {
	ev := &emitter.Event{Args: []any{"only one"}}

	// Arg method should return nil for out of bounds
	if ev.Arg(5) != nil {
		t.Error("expected nil for out of bounds Arg")
	}
}

func TestEventArgEmpty(t *testing.T) {
	ev := &emitter.Event{Args: []any{}}

	if ev.Arg(0) != nil {
		t.Error("expected nil for empty args")
	}
}

func TestEventArgNil(t *testing.T) {
	ev := &emitter.Event{}

	if ev.Arg(0) != nil {
		t.Error("expected nil for nil args")
	}
}

func TestEncodedArgCaching(t *testing.T) {
	ev := &emitter.Event{Args: []any{map[string]int{"count": 42}}}

	callCount := 0
	encoder := func(v any) ([]byte, error) {
		callCount++
		return json.Marshal(v)
	}

	// First call
	result1, err := ev.EncodedArg(0, "json", encoder)
	if err != nil {
		t.Errorf("first encode failed: %v", err)
	}

	// Second call - should use cache
	result2, err := ev.EncodedArg(0, "json", encoder)
	if err != nil {
		t.Errorf("second encode failed: %v", err)
	}

	if callCount != 1 {
		t.Errorf("encoder called %d times, expected 1 (caching failed)", callCount)
	}

	if string(result1) != string(result2) {
		t.Error("cached result differs from original")
	}
}

func TestEncodedArgDifferentKeys(t *testing.T) {
	ev := &emitter.Event{Args: []any{"test"}}

	jsonCallCount := 0
	jsonEncoder := func(v any) ([]byte, error) {
		jsonCallCount++
		return json.Marshal(v)
	}

	customCallCount := 0
	customEncoder := func(v any) ([]byte, error) {
		customCallCount++
		return []byte("custom:" + v.(string)), nil
	}

	// Call with different keys
	jsonResult, _ := ev.EncodedArg(0, "json", jsonEncoder)
	customResult, _ := ev.EncodedArg(0, "custom", customEncoder)

	// Both should have been called once
	if jsonCallCount != 1 {
		t.Errorf("json encoder called %d times", jsonCallCount)
	}
	if customCallCount != 1 {
		t.Errorf("custom encoder called %d times", customCallCount)
	}

	if string(jsonResult) != `"test"` {
		t.Errorf("unexpected json result: %s", jsonResult)
	}
	if string(customResult) != "custom:test" {
		t.Errorf("unexpected custom result: %s", customResult)
	}
}

func TestEncodedArgInvalidIndex(t *testing.T) {
	ev := &emitter.Event{Args: []any{"only one"}}

	_, err := ev.EncodedArg(5, "json", json.Marshal)
	if err == nil {
		t.Error("expected error for invalid index")
	}
}

func TestEncodedArgEncoderError(t *testing.T) {
	ev := &emitter.Event{Args: []any{"test"}}

	expectedErr := errors.New("encoder error")
	failingEncoder := func(v any) ([]byte, error) {
		return nil, expectedErr
	}

	_, err := ev.EncodedArg(0, "failing", failingEncoder)
	if err == nil {
		t.Error("expected encoder error")
	}

	// Second call should return same cached error
	_, err2 := ev.EncodedArg(0, "failing", failingEncoder)
	if err2 == nil {
		t.Error("expected cached encoder error")
	}
}

func TestEncodedArgMultipleArgs(t *testing.T) {
	ev := &emitter.Event{Args: []any{"first", "second", "third"}}

	r0, _ := ev.EncodedArg(0, "json", json.Marshal)
	r1, _ := ev.EncodedArg(1, "json", json.Marshal)
	r2, _ := ev.EncodedArg(2, "json", json.Marshal)

	if string(r0) != `"first"` {
		t.Errorf("unexpected r0: %s", r0)
	}
	if string(r1) != `"second"` {
		t.Errorf("unexpected r1: %s", r1)
	}
	if string(r2) != `"third"` {
		t.Errorf("unexpected r2: %s", r2)
	}
}

func TestEventFields(t *testing.T) {
	ev := &emitter.Event{
		Topic: "test-topic",
		Args:  []any{1, 2, 3},
	}

	if ev.Topic != "test-topic" {
		t.Errorf("unexpected topic: %s", ev.Topic)
	}

	if len(ev.Args) != 3 {
		t.Errorf("unexpected args length: %d", len(ev.Args))
	}
}
