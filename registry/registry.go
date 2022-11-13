package registry

import (
	"context"
	"errors"
)

// The Registry provides an interface for service discovery
// and an abstraction over varying implementations
// {consul, etcd, zookeeper, ...}
type Registry interface {
	Register(ctx context.Context, service Service, options ...WithRegisterOption) error
	Deregister(ctx context.Context, service Service) error
	GetService(ctx context.Context, serviceName string) ([]Service, error)
	ListServices(ctx context.Context) ([]Service, error)
	Watch(ctx context.Context, serviceName string) (Watcher, error)
}

var (
	// ErrNotFound Not found error when GetService is called
	ErrNotFound = errors.New("service not found")
	// ErrWatcherStopped Watcher stopped error when watcher is stopped
	ErrWatcherStopped = errors.New("watcher stopped")
)

type Service struct {
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Metadata  map[string]string `json:"metadata"`
	Endpoints []Endpoint        `json:"endpoints"`
	Nodes     []Node            `json:"nodes"`
}

type Endpoint struct {
	Name     string            `json:"name"`
	Request  *Value            `json:"request"`
	Response *Value            `json:"response"`
	Metadata map[string]string `json:"metadata"`
}

type Value struct {
	Name   string  `json:"name"`
	Type   string  `json:"type"`
	Values []Value `json:"values"`
}

type Node struct {
	Id       string            `json:"id"`
	Address  string            `json:"address"`
	Metadata map[string]string `json:"metadata"`
}
