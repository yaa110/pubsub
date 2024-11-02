package pubsub_test

import (
	"sync"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/yaa110/pubsub"
)

func TestPubSub1(t *testing.T) {
	pubsub.Isolate(t)
	assert := gomega.NewWithT(t)

	chn := make(chan *int, 1)
	recv := &dummyReceiver[int]{}
	err := pubsub.Subscribe[int](chn)
	assert.Expect(err).NotTo(gomega.HaveOccurred())
	err = pubsub.Subscribe[int](recv)
	assert.Expect(err).NotTo(gomega.HaveOccurred())

	pubsub.Publish(pointer(1))

	assert.Expect(*<-chn).To(gomega.Equal(1))
	assert.Eventually(func() int {
		return recv.Read()
	}, 10*time.Second, 250*time.Millisecond).Should(gomega.Equal(1))
}

func TestPubSub2(t *testing.T) {
	pubsub.Isolate(t)
	assert := gomega.NewWithT(t)

	chn := make(chan *int, 1)
	recv := &dummyReceiver[int]{}
	err := pubsub.Subscribe[int](chn)
	assert.Expect(err).NotTo(gomega.HaveOccurred())
	err = pubsub.Subscribe[int](recv)
	assert.Expect(err).NotTo(gomega.HaveOccurred())

	pubsub.Publish(pointer(3))

	assert.Expect(*<-chn).To(gomega.Equal(3))
	assert.Eventually(func() int {
		return recv.Read()
	}, 10*time.Second, 250*time.Millisecond).Should(gomega.Equal(3))
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
