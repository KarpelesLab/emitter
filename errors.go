package emitter

import "errors"

// ErrNoSuchTopic is returned by [Hub.Emit] and [Hub.EmitEvent] when attempting
// to emit an event to a topic that has no subscribers.
var ErrNoSuchTopic = errors.New("no such topic")
