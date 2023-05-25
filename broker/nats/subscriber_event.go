package nats

import (
	"errors"
	"github.com/nats-io/nats.go"
	"strings"
)

type _NatsEvent struct {
	msg *nats.Msg
	ns  *_NatsSubscriber
}

// Pattern returns the subscription pattern used to create the event.
func (e _NatsEvent) Pattern() string {
	return e.ns.Pattern()
}

// Topic returns the topic the event was received on.
func (e _NatsEvent) Topic() string {
	return strings.TrimPrefix(e.msg.Subject, e.ns.nb.options.TopicPrefix)
}

// QueueName subscribers with the same queue name will create a shared subscription where each receives a subset of messages.
func (e _NatsEvent) QueueName() string {
	return e.ns.QueueName()
}

// Message returns the raw message payload of the event.
func (e _NatsEvent) Message() []byte {
	return e.msg.Data
}

// Ack acknowledges the successful processing of the event. It indicates that the event can be removed from the subscription queue.
func (e _NatsEvent) Ack() error {
	return errors.New("not using JetStream, unable to acknowledge(ack)")
}

// Error returns any error that occurred while processing the event, if applicable.
func (e _NatsEvent) Error() error {
	return nil
}
