package pubsub

import (
	"fmt"
	"reflect"
	"sync"
)

type pubSub struct {
	subscribers map[string]map[uint64]SubscribeDescriptor
	lock        sync.RWMutex
	lastID      uint64
	freeIDs     []uint64
}

type SubscribeDescriptor interface {
	message
	receive(i interface {
		message
		intoInner
	})
	setID(id uint64)
	ID() uint64
}

type receiver[T any, PT interface{ *T }] struct {
	receiver Receiver[T, PT]
	id       uint64
}

type channel[T any, PT interface{ *T }] struct {
	chn chan PT
	id  uint64
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
		subscribers: make(map[string]map[uint64]SubscribeDescriptor),
	}
}

// newSubscriber converts a channel or a receiver to a subscriber to receive incoming messages.
func newSubscriber[T any, PT interface{ *T }](sub any) (SubscribeDescriptor, error) {
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

func (p *pubSub) allocateID() uint64 {
	var id uint64
	if len(p.freeIDs) > 0 {
		id, p.freeIDs = p.freeIDs[0], p.freeIDs[1:]
	} else {
		id = p.lastID
		p.lastID++
	}
	return id
}

func (p *pubSub) subscribe(sd SubscribeDescriptor) {
	topic := sd.topic()

	p.lock.Lock()
	defer p.lock.Unlock()

	id := p.allocateID()
	if _, ok := p.subscribers[topic]; !ok {
		p.subscribers[topic] = make(map[uint64]SubscribeDescriptor)
	}
	p.subscribers[topic][id] = sd
}

func (p *pubSub) unsubscribe(sd SubscribeDescriptor) {
	topic := sd.topic()

	p.lock.Lock()
	defer p.lock.Unlock()

	p.freeIDs = append(p.freeIDs, sd.ID())
	if _, ok := p.subscribers[topic]; ok {
		delete(p.subscribers[topic], sd.ID())
	}
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

func (w *receiver[T, PT]) ID() uint64 {
	return w.id
}

func (w *channel[T, PT]) ID() uint64 {
	return w.id
}

func (w *receiver[T, PT]) setID(id uint64) {
	w.id = id
}

func (w *channel[T, PT]) setID(id uint64) {
	w.id = id
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
