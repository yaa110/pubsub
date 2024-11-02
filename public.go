package pubsub

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"sync"
)

var (
	// global is the default pubsub to be used for messaging.
	global     *pubSub
	isolations sync.Map
)

const isolateEnv = "PUBSUB_ISOLATE_SUBSYSTEM"

type Receiver[T any, PT interface{ *T }] interface {
	Receive(ptr PT)
}

func init() {
	global = create()
}

// Isolate sets up a separate PubSub instance to allow tests to run concurrently.
func Isolate[T EnvSetter](t T) {
	global = nil
	uuid := uuid4()
	t.Setenv(isolateEnv, uuid)
	isolations.Store(uuid, create())
}

// Publish publishes message `m` to a PubSub instance.
// This method is thread-safe.
func Publish[T any](m *T) {
	if global != nil {
		global.publish(newMessage(m))
	} else if name := os.Getenv(isolateEnv); name != "" {
		if ps, ok := isolations.Load(name); ok {
			ps.(*pubSub).publish(newMessage(m))
		}
	}
}

// Subscribe subscribes `sub` to a PubSub instance.
// `sub` can be a `pubsub.Receiver` or `chan *T`.
// This method is thread-safe.
func Subscribe[T any, PT interface{ *T }](sub any) error {
	subscriber, err := newSubscriber[T](sub)
	if err != nil {
		return err
	}
	if global != nil {
		global.subscribe(subscriber)
	} else if name := os.Getenv(isolateEnv); name != "" {
		if ps, ok := isolations.Load(name); ok {
			ps.(*pubSub).subscribe(subscriber)
		}
	}
	return nil
}

func uuid4() string {
	uuid := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, uuid); err != nil {
		panic(err)
	}
	uuid[6] = (uuid[6] & 0x0F) | 0x40
	uuid[8] = (uuid[8] & 0x3F) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

type EnvSetter interface {
	Setenv(key, val string)
}
