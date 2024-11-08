# PubSub for Golang

[![Test](https://github.com/yaa110/pubsub/actions/workflows/test.yaml/badge.svg)](https://github.com/yaa110/pubsub/actions/workflows/test.yaml) [![goreportcard](https://img.shields.io/badge/go%20report-A%2B-brightgreen.svg)](http://goreportcard.com/report/yaa110/pubsub) [![License](http://img.shields.io/:license-mit-blue.svg)](https://github.com/yaa110/pubsub/blob/master/LICENSE) [![codecov](https://codecov.io/github/yaa110/pubsub/graph/badge.svg?token=UV0HRPA0C4)](https://codecov.io/github/yaa110/pubsub) [![godoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/yaa110/pubsub)

This package provides an internal PubSub system for Golang, allowing messages to be published and subscribed to.

## Getting Started

- Import package:

```go
import "github.com/yaa110/pubsub"
```

- The pubsub package supports two types of subscribers: channels and types that implement the `pubsub.Receiver` interface. For instance, you can define a receiver as follows:

```go
type CustomMessage struct {
    Content string
}

type CustomReceiver struct {
}

func (d *CustomReceiver) Receive(msg *CustomMessage) {
    slog.Info("received message", "content", msg.Content)
}
```

- Subscribe/Unsubscribe for a type of message:

```go
// Using a channel as subscriber.
chn := make(chan *CustomMessage)
sd1 := pubsub.Subscribe[CustomMessage](chn) // returns a descriptor

// Using a custom receiver as subscriber.
receiver := &CustomReceiver{}
sd2 := pubsub.Subscribe[CustomMessage](receiver)

// Unsubscribe using the descriptor.
pubsub.Unsubscribe(sd1)
```

**Note** that a topic is automatically created for each message type, allowing all subscribers of that message type to receive the published messages.

- Publish a message:

```go
pubsub.Publish(&CustomMessage{
    Content: "data",
})
```

**Note** that all subscribers will receive a pointer to the published message, with each receiver running in a new goroutine, so the user should handle concurrent calls to the `Receive` method.

The pubsub package can also be used with built-in types like `string` or `int`; you simply need to pass a pointer to these types:

```go
chn := make(chan *int, 1) // also a custom receiver can be used to receive *int values
pubsub.Subscribe[int](chn)
assert(*<-chn, 2)

pubsub.Publish(pointer(2)) // publishes to "int" topic

func pointer[T any](t T) *T {
    return &t
}
```

**Note** that published messages will only be received by subscribers that were subscribed prior to the message being published.

A single instance of pubsub can handle publishing and subscribing for multiple message types:

```go
pubsub.Publish(pointer(2))      // publishes to "int" topic
pubsub.Publish(pointer("test")) // publishes to "string" topic
```

## Test Isolation

Since a global pubsub instance is created for the entire program, it can lead to issues when tests are run concurrently. To resolve this, call `pubsub.Isolate` at the beginning of each test:

```go
func TestIsolatedPubSub(t *testing.T) {
    pubsub.Isolate(t)
    // ...
}

func BenchmarkIsolatedPubSub(b *testing.B) {
    pubsub.Isolate(b)
    // ...
}
```
