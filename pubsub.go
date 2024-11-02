package pubsub

import (
	"fmt"
	"reflect"
	"sync"
)

type pubSub struct {
	subscribers map[string][]subscriber
	lock        sync.RWMutex
}

type subscriber interface {
	message
	receive(i interface {
		message
		intoInner
	})
}

type receiver[T any, PT interface{ *T }] struct {
	receiver Receiver[T, PT]
}

type channel[T any, PT interface{ *T }] struct {
	chn chan PT
}

type message interface {
	topic() string
}

type intoInner interface {
	inner() any
}

type msg[T any] struct {
	t *T
}

// create creates a create instance of pubsub to publish messages and receive them.
func create() *pubSub {
	return &pubSub{
		subscribers: make(map[string][]subscriber),
	}
}

// newSubscriber converts a channel or a receiver to a subscriber to receive incoming messages.
func newSubscriber[T any, PT interface{ *T }](sub any) (subscriber, error) {
	switch s := sub.(type) {
	case chan PT:
		return &channel[T, PT]{
			chn: s,
		}, nil
	case Receiver[T, PT]:
		return &receiver[T, PT]{
			receiver: s,
		}, nil
	default:
		return nil, fmt.Errorf("unknown subscriber, expected: Recever[T, PT] or chan PT, got: %s", reflect.TypeOf(s))
	}
}

func (p *pubSub) subscribe(sub subscriber) {
	topic := sub.topic()

	p.lock.Lock()
	defer p.lock.Unlock()

	p.subscribers[topic] = append(p.subscribers[topic], sub)
}

func (p *pubSub) publish(m interface {
	message
	intoInner
}) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	for _, subscriber := range p.subscribers[m.topic()] {
		go subscriber.receive(m)
	}
}

func (w *receiver[T, PT]) receive(m interface {
	message
	intoInner
}) {
	w.receiver.Receive(m.inner().(PT))
}

func (w *channel[T, PT]) receive(m interface {
	message
	intoInner
}) {
	w.chn <- m.inner().(PT)
}

func (w *receiver[T, PT]) topic() string {
	var zero T
	return reflect.TypeOf(zero).String()
}

func (w *channel[T, PT]) topic() string {
	var zero T
	return reflect.TypeOf(zero).String()
}

func (m *msg[T]) topic() string {
	var zero T
	return reflect.TypeOf(zero).String()
}

func (m *msg[T]) inner() any {
	return m.t
}

func newMessage[T any](t *T) *msg[T] {
	return &msg[T]{
		t: t,
	}
}
