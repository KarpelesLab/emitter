package emitter_test

import (
	"context"
	"log"
	"testing"

	"github.com/KarpelesLab/emitter"
)

func TestEvents(t *testing.T) {
	h := emitter.New()

	go handleTest1(h.On("test1"))

	h.Emit(context.Background(), "test1", "hello world")
}

func handleTest1(ch <-chan *emitter.Event) {
	for ev := range ch {
		log.Printf("ev = %v", ev)
	}
}
