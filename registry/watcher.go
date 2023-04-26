package registry

import (
	"errors"
	"fmt"
	"time"
)

// Watcher is an interface that returns updates
// about services within the registry.
type Watcher interface {
	// Next is a blocking call
	Next() (*Result, error)
	// Stop stop watching
	Stop()
}

// Result is returned by a call to Next on
// the watcher. Actions can be create, update, delete
type Result struct {
	Action  string
	Service *Service
}

// EventType defines registry event type
type EventType int

const (
	// Create is emitted when a new service is registered
	Create EventType = iota
	// Delete is emitted when an existing service is deregsitered
	Delete
	// Update is emitted when an existing service is updated
	Update
)

// String returns human readable event type
func (t EventType) String() string {
	switch t {
	case Create:
		return "create"
	case Delete:
		return "delete"
	case Update:
		return "update"
	default:
		return "unknown"
	}
}

// Set converts a EventType string into a EventType value.
// returns error if the input string does not match known values.
func (t *EventType) Set(str string) error {
	if t == nil {
		return errors.New("can't set a nil *EventType")
	}

	switch str {
	case Create.String():
		*t = Create
	case Delete.String():
		*t = Delete
	case Update.String():
		*t = Update
	}

	return fmt.Errorf("unrecognized EventType: %q", str)
}

// Event is registry event
type Event struct {
	// Id is registry id
	Id string `json:"id"`
	// Type defines type of event
	Type EventType `json:"type"`
	// Timestamp is event timestamp
	Timestamp time.Time `json:"ts"`
	// Service is registry service
	Service *Service `json:"service"`
}
