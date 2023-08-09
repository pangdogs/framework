package redis

import (
	"github.com/redis/go-redis/v9"
	"net"
	"strings"
)

type Option struct{}

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

func (Option) Default() RegistryOption {
	return func(options *RegistryOptions) {
		Option{}.RedisClient(nil)(options)
		Option{}.RedisConfig(nil)(options)
		Option{}.RedisURL("")(options)
		Option{}.KeyPrefix("golaxy:registry:")(options)
		Option{}.WatchChanSize(128)(options)
		Option{}.FastAuth("", "")(options)
		Option{}.FastAddress("127.0.0.1:6379")(options)
		Option{}.FastDBIndex(0)(options)
	}
}

func (Option) RedisClient(cli *redis.Client) RegistryOption {
	return func(o *RegistryOptions) {
		o.RedisClient = cli
	}
}

func (Option) RedisConfig(conf *redis.Options) RegistryOption {
	return func(o *RegistryOptions) {
		o.RedisConfig = conf
	}
}

func (Option) RedisURL(url string) RegistryOption {
	return func(o *RegistryOptions) {
		o.RedisURL = url
	}
}

func (Option) KeyPrefix(prefix string) RegistryOption {
	return func(o *RegistryOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, ":") {
			prefix += ":"
		}
		o.KeyPrefix = prefix
	}
}

func (Option) WatchChanSize(size int) RegistryOption {
	return func(o *RegistryOptions) {
		if size < 0 {
			panic("option WatchChanSize can't be set to a value less then 0")
		}
		o.WatchChanSize = size
	}
}

func (Option) FastAuth(username, password string) RegistryOption {
	return func(options *RegistryOptions) {
		options.FastUsername = username
		options.FastPassword = password
	}
}

func (Option) FastAddress(addr string) RegistryOption {
	return func(options *RegistryOptions) {
		if _, _, err := net.SplitHostPort(addr); err != nil {
			panic(err)
		}
		options.FastAddress = addr
	}
}

func (Option) FastDBIndex(idx int) RegistryOption {
	return func(options *RegistryOptions) {
		options.FastDBIndex = idx
	}
}
