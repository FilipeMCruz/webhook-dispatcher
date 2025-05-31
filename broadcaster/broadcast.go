package broadcaster

import (
	"context"
	"github.com/google/uuid"
)

type BroadcastServer[T any] interface {
	Send(T)
	Subscribe(uuid.UUID, int) <-chan T
	Unsubscribe(uuid.UUID)
}

type addListenerCommand[T any] struct {
	id uuid.UUID
	ch chan T
}

type broadcastServer[T any] struct {
	source         chan T
	listeners      map[uuid.UUID]chan T
	addListener    chan addListenerCommand[T]
	removeListener chan uuid.UUID
}

func (s *broadcastServer[T]) Send(t T) {
	s.source <- t
}

func (s *broadcastServer[T]) Subscribe(id uuid.UUID, bufferSize int) <-chan T {
	newListener := make(chan T, bufferSize)
	s.addListener <- addListenerCommand[T]{
		id: id,
		ch: newListener,
	}
	return newListener
}

func (s *broadcastServer[T]) Unsubscribe(id uuid.UUID) {
	s.removeListener <- id
}

func (s *broadcastServer[T]) serve(ctx context.Context) {
	defer func() {
		for _, listener := range s.listeners {
			if listener != nil {
				close(listener)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case newListener := <-s.addListener:
			s.listeners[newListener.id] = newListener.ch
		case listenerToRemove := <-s.removeListener:
			ch, ok := s.listeners[listenerToRemove]
			if ok {
				delete(s.listeners, listenerToRemove)
				close(ch)
			}
		case val, ok := <-s.source:
			if !ok {
				return
			}
			for _, listener := range s.listeners {
				if listener != nil {
					select {
					case listener <- val:
					case <-ctx.Done():
						return
					}

				}
			}
		}
	}
}

func NewBroadcastServer[T any](ctx context.Context) BroadcastServer[T] {
	service := &broadcastServer[T]{
		source:         make(chan T),
		listeners:      make(map[uuid.UUID]chan T),
		addListener:    make(chan addListenerCommand[T]),
		removeListener: make(chan uuid.UUID),
	}
	go service.serve(ctx)

	return service
}
