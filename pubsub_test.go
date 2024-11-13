package pubsub_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/yaa110/pubsub"
)

func TestPubSub(t *testing.T) {
	for _, isolate := range []bool{false, true} {
		if isolate {
			pubsub.Isolate(t)
		}

		assert := gomega.NewWithT(t)

		chn := make(chan *int, 1)
		recv := &dummyReceiver[int]{}
		chnSD := pubsub.Subscribe[int](chn)
		assert.Expect(chnSD).NotTo(gomega.BeNil())
		recvSD := pubsub.Subscribe[int](recv)
		assert.Expect(recvSD).NotTo(gomega.BeNil())

		pubsub.Publish(pointer(1))

		assert.Expect(*<-chn).To(gomega.Equal(1))
		assert.Eventually(func() int {
			return recv.Read()
		}, 10*time.Second, 250*time.Millisecond).Should(gomega.Equal(1))

		pubsub.Unsubscribe(chnSD)
		pubsub.Publish(pointer(2))

		assert.Eventually(func() int {
			return recv.Read()
		}, 10*time.Second, 250*time.Millisecond).Should(gomega.Equal(2))

		assert.Expect(func() error {
			select {
			case <-chn:
				return errors.New("unsubscribe failed")
			case <-time.After(time.Second):
				return nil
			}
		}()).Should(gomega.Succeed())
	}
}

func TestPubSubDescriptor(t *testing.T) {
	assert := gomega.NewWithT(t)
	c := make(chan *int)
	for i := 0; i < 1000; i++ {
		sd := pubsub.Subscribe[int](c)
		assert.Expect(sd.ID()).To(gomega.Equal(uint64(0)))
		pubsub.Unsubscribe(sd)
	}
}

type dummyReceiver[T any] struct {
	msg   T
	mutex sync.Mutex
}

func (d *dummyReceiver[T]) Read() T {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.msg
}

func (d *dummyReceiver[T]) Receive(msg *T) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.msg = *msg
}

func pointer[T any](t T) *T {
	return &t
}
