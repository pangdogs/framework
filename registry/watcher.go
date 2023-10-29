package registry

import (
	"bytes"
	"fmt"
)

// Watcher is an interface that returns updates
// about services within the registry.
type Watcher interface {
	// Next is a blocking call
	Next() (*Event, error)
	// Stop stop watching
	Stop()
}

// Event is returned by a call to Next on the watcher. Type can be create, update, delete
// +k8s:deepcopy-gen=true
type Event struct {
	Type    EventType `json:"type"`
	Service *Service  `json:"service"`
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

// MarshalText marshals the EventType to text.
func (t *EventType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText unmarshals text to a EventType.
func (t *EventType) UnmarshalText(text []byte) error {
	if t == nil {
		return fmt.Errorf("%w: can't unmarshal a nil *EventType", ErrRegistry)
	}
	if !t.unmarshalText(text) && !t.unmarshalText(bytes.ToLower(text)) {
		return fmt.Errorf("%w: unrecognized EventType: %q", ErrRegistry, text)
	}
	return nil
}

func (t *EventType) unmarshalText(text []byte) bool {
	switch string(text) {
	case "create", "CREATE":
		*t = Create
	case "delete", "DELETE":
		*t = Delete
	case "update", "UPDATE":
		*t = Update
	default:
		return false
	}
	return true
}

// String returns human readable EventType.
func (t EventType) String() string {
	switch t {
	case Create:
		return "create"
	case Delete:
		return "delete"
	case Update:
		return "update"
	default:
		return fmt.Sprintf("EventType(%d)", t)
	}
}

// Set converts a EventType string into a EventType value.
// returns error if the input string does not match known values.
func (t *EventType) Set(str string) error {
	return t.UnmarshalText([]byte(str))
}
