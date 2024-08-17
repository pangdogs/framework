/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package discovery

import (
	"bytes"
	"errors"
	"fmt"
)

// IWatcher is an interface that returns updates
// about services within the registry.
type IWatcher interface {
	// Pattern watching pattern
	Pattern() string
	// Next is a blocking call
	Next() (*Event, error)
	// Terminate stop watching
	Terminate() <-chan struct{}
	// Terminated stopped notify
	Terminated() <-chan struct{}
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
		return errors.New("registry: can't unmarshal a nil *EventType")
	}
	if !t.unmarshalText(text) && !t.unmarshalText(bytes.ToLower(text)) {
		return fmt.Errorf("registry: unrecognized EventType: %q", text)
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
