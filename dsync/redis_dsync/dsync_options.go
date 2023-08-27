package redis_dsync

import (
	"github.com/redis/go-redis/v9"
	"net"
	"strings"
)

type Option struct{}

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

func (Option) Default() DSyncOption {
	return func(options *DSyncOptions) {
		Option{}.RedisClient(nil)(options)
		Option{}.RedisConfig(nil)(options)
		Option{}.RedisURL("")(options)
		Option{}.KeyPrefix("golaxy:mutex:")(options)
		Option{}.FastAuth("", "")(options)
		Option{}.FastAddress("127.0.0.1:6379")(options)
		Option{}.FastDBIndex(0)(options)
	}
}

func (Option) RedisClient(cli *redis.Client) DSyncOption {
	return func(o *DSyncOptions) {
		o.RedisClient = cli
	}
}

func (Option) RedisConfig(conf *redis.Options) DSyncOption {
	return func(o *DSyncOptions) {
		o.RedisConfig = conf
	}
}

func (Option) RedisURL(url string) DSyncOption {
	return func(o *DSyncOptions) {
		o.RedisURL = url
	}
}

func (Option) KeyPrefix(prefix string) DSyncOption {
	return func(o *DSyncOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, ":") {
			prefix += ":"
		}
		o.KeyPrefix = prefix
	}
}

func (Option) FastAuth(username, password string) DSyncOption {
	return func(options *DSyncOptions) {
		options.FastUsername = username
		options.FastPassword = password
	}
}

func (Option) FastAddress(addr string) DSyncOption {
	return func(options *DSyncOptions) {
		if _, _, err := net.SplitHostPort(addr); err != nil {
			panic(err)
		}
		options.FastAddress = addr
	}
}

func (Option) FastDBIndex(idx int) DSyncOption {
	return func(options *DSyncOptions) {
		options.FastDBIndex = idx
	}
}
