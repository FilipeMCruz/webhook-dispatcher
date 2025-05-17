package broadcaster

import (
	"context"
	"github.com/google/uuid"
	"testing"
)

func TestNewBroadcastServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan int)

	s := NewBroadcastServer(ctx, in)

	sid1 := uuid.New()
	ch1 := s.Subscribe(sid1)
	defer s.Unsubscribe(sid1)

	sid2 := uuid.New()
	ch2 := s.Subscribe(sid2)
	defer s.Unsubscribe(sid2)

	v1 := 1
	in <- v1
	r1 := <-ch1
	r2 := <-ch2

	if r1 != v1 {
		t.Errorf("got %d, want %d", r1, v1)
	}

	if r2 != v1 {
		t.Errorf("got %d, want %d", r2, v1)
	}
}
