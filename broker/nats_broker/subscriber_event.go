package nats_broker

import (
	"context"
	"errors"
	"github.com/nats-io/nats.go"
	"strings"
)

type _Event struct {
	msg *nats.Msg
	ns  *_Subscriber
}

// Pattern returns the subscription pattern used to create the event.
func (e *_Event) Pattern() string {
	return e.ns.Pattern()
}

// Topic returns the topic the event was received on.
func (e *_Event) Topic() string {
	return strings.TrimPrefix(e.msg.Subject, e.ns.broker.options.TopicPrefix)
}

// Queue subscribers with the same queue name will create a shared subscription where each receives a subset of messages.
func (e *_Event) Queue() string {
	return e.ns.Queue()
}

// Message returns the raw message payload of the event.
func (e *_Event) Message() []byte {
	return e.msg.Data
}

// Ack acknowledges the successful processing of the event. It indicates that the event can be removed from the subscription queue.
func (e *_Event) Ack(ctx context.Context) error {
	return errors.New("used not JetStream, unable to acknowledge(ack)")
}

// Nak negatively acknowledges a message. This tells the server to redeliver the message.
func (e *_Event) Nak(ctx context.Context) error {
	return errors.New("used not JetStream, unable to negatively acknowledge(nak)")
}
