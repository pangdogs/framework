package redis

import (
	"github.com/redis/go-redis/v9"
	"net"
	"strings"
)

type WithOption struct{}

type DSyncOptions struct {
	RedisClient  *redis.Client
	RedisConfig  *redis.Options
	RedisURL     string
	KeyPrefix    string
	FastUsername string
	FastPassword string
	FastAddress  string
	FastDBIndex  int
}

type DSyncOption func(options *DSyncOptions)

func (WithOption) Default() DSyncOption {
	return func(options *DSyncOptions) {
		WithOption{}.RedisClient(nil)(options)
		WithOption{}.RedisConfig(nil)(options)
		WithOption{}.RedisURL("")(options)
		WithOption{}.KeyPrefix("golaxy:mutex:")(options)
		WithOption{}.FastAuth("", "")(options)
		WithOption{}.FastAddress("127.0.0.1:6379")(options)
		WithOption{}.FastDBIndex(0)(options)
	}
}

func (WithOption) RedisClient(cli *redis.Client) DSyncOption {
	return func(o *DSyncOptions) {
		o.RedisClient = cli
	}
}

func (WithOption) RedisConfig(conf *redis.Options) DSyncOption {
	return func(o *DSyncOptions) {
		o.RedisConfig = conf
	}
}

func (WithOption) RedisURL(url string) DSyncOption {
	return func(o *DSyncOptions) {
		o.RedisURL = url
	}
}

func (WithOption) KeyPrefix(prefix string) DSyncOption {
	return func(o *DSyncOptions) {
		if !strings.HasSuffix(prefix, ":") {
			prefix += ":"
		}
		o.KeyPrefix = prefix
	}
}

func (WithOption) FastAuth(username, password string) DSyncOption {
	return func(options *DSyncOptions) {
		options.FastUsername = username
		options.FastPassword = password
	}
}

func (WithOption) FastAddress(addr string) DSyncOption {
	return func(options *DSyncOptions) {
		if _, _, err := net.SplitHostPort(addr); err != nil {
			panic(err)
		}
		options.FastAddress = addr
	}
}

func (WithOption) FastDBIndex(idx int) DSyncOption {
	return func(options *DSyncOptions) {
		options.FastDBIndex = idx
	}
}
