package redis

import (
	"github.com/redis/go-redis/v9"
	"net"
	"strings"
)

type Options struct {
	RedisClient  *redis.Client
	RedisConfig  *redis.Options
	RedisURL     string
	KeyPrefix    string
	FastUsername string
	FastPassword string
	FastAddress  string
	FastDBIndex  int
}

type Option func(options *Options)

type WithOption struct{}

func (WithOption) Default() Option {
	return func(options *Options) {
		WithOption{}.RedisClient(nil)(options)
		WithOption{}.RedisConfig(nil)(options)
		WithOption{}.RedisURL("")(options)
		WithOption{}.KeyPrefix("golaxy:mutex:")(options)
		WithOption{}.FastAuth("", "")(options)
		WithOption{}.FastAddress("127.0.0.1:6379")(options)
		WithOption{}.FastDBIndex(0)(options)
	}
}

func (WithOption) RedisClient(cli *redis.Client) Option {
	return func(o *Options) {
		o.RedisClient = cli
	}
}

func (WithOption) RedisConfig(conf *redis.Options) Option {
	return func(o *Options) {
		o.RedisConfig = conf
	}
}

func (WithOption) RedisURL(url string) Option {
	return func(o *Options) {
		o.RedisURL = url
	}
}

func (WithOption) KeyPrefix(prefix string) Option {
	return func(o *Options) {
		if !strings.HasSuffix(prefix, ":") {
			prefix += ":"
		}
		o.KeyPrefix = prefix
	}
}

func (WithOption) FastAuth(username, password string) Option {
	return func(options *Options) {
		options.FastUsername = username
		options.FastPassword = password
	}
}

func (WithOption) FastAddress(addr string) Option {
	return func(options *Options) {
		if _, _, err := net.SplitHostPort(addr); err != nil {
			panic(err)
		}
		options.FastAddress = addr
	}
}

func (WithOption) FastDBIndex(idx int) Option {
	return func(options *Options) {
		options.FastDBIndex = idx
	}
}
