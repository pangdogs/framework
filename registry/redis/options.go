package redis

import (
	"github.com/redis/go-redis/v9"
	"net"
	"strings"
)

type WithOption struct{}

type RegistryOptions struct {
	RedisClient   *redis.Client
	RedisConfig   *redis.Options
	RedisURL      string
	KeyPrefix     string
	WatchChanSize int
	FastUsername  string
	FastPassword  string
	FastAddress   string
	FastDBIndex   int
}

type RegistryOption func(options *RegistryOptions)

func (WithOption) Default() RegistryOption {
	return func(options *RegistryOptions) {
		WithOption{}.RedisClient(nil)(options)
		WithOption{}.RedisConfig(nil)(options)
		WithOption{}.RedisURL("")(options)
		WithOption{}.KeyPrefix("golaxy:registry:")(options)
		WithOption{}.WatchChanSize(128)(options)
		WithOption{}.FastAuth("", "")(options)
		WithOption{}.FastAddress("127.0.0.1:6379")(options)
		WithOption{}.FastDBIndex(0)(options)
	}
}

func (WithOption) RedisClient(cli *redis.Client) RegistryOption {
	return func(o *RegistryOptions) {
		o.RedisClient = cli
	}
}

func (WithOption) RedisConfig(conf *redis.Options) RegistryOption {
	return func(o *RegistryOptions) {
		o.RedisConfig = conf
	}
}

func (WithOption) RedisURL(url string) RegistryOption {
	return func(o *RegistryOptions) {
		o.RedisURL = url
	}
}

func (WithOption) KeyPrefix(prefix string) RegistryOption {
	return func(o *RegistryOptions) {
		if !strings.HasSuffix(prefix, ":") {
			prefix += ":"
		}
		o.KeyPrefix = prefix
	}
}

func (WithOption) WatchChanSize(size int) RegistryOption {
	return func(o *RegistryOptions) {
		if size < 0 {
			panic("option WatchChanSize can't be set to a value less then 0")
		}
		o.WatchChanSize = size
	}
}

func (WithOption) FastAuth(username, password string) RegistryOption {
	return func(options *RegistryOptions) {
		options.FastUsername = username
		options.FastPassword = password
	}
}

func (WithOption) FastAddress(addr string) RegistryOption {
	return func(options *RegistryOptions) {
		if _, _, err := net.SplitHostPort(addr); err != nil {
			panic(err)
		}
		options.FastAddress = addr
	}
}

func (WithOption) FastDBIndex(idx int) RegistryOption {
	return func(options *RegistryOptions) {
		options.FastDBIndex = idx
	}
}
