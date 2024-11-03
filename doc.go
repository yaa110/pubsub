// Package pubsub provides an internal PubSub system, allowing messages to be published and subscribed to.
//
// Example usage:
//
//	chn := make(chan *int, 1)
//	sd := pubsub.Subscribe[int](chn)
//	pubsub.Publish(pointer(1))
//	assert(*<-chn, 1)
//	pubsub.Unsubscribe(sd)
//
// For additional examples and receivers, please refer to the README file.
package pubsub
