package distributed

import (
	"kit.golaxy.org/golaxy/service"
	"sync"
)

type Distributed interface {
}

func newDistributed(options ...any) Distributed {
	opts := RegistryOptions{}
	Option{}.Default()(&opts)

	for i := range options {
		options[i](&opts)
	}

	return &_RedisRegistry{
		options:  opts,
		register: map[string]uint64{},
	}
}

type _Distributed struct {
	options  RegistryOptions
	ctx      service.Context
	client   *redis.Client
	register map[string]uint64
	mutex    sync.RWMutex
}
