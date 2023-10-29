package redis_dsync

import (
	"github.com/redis/go-redis/v9"
	"net"
	"strings"
)

// Option is a struct used for setting options.
type Option struct{}

// DSyncOptions contains various options for configuring distributed locking using redis.
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

// DSyncOption is a function type for configuring DSyncOptions.
type DSyncOption func(options *DSyncOptions)

// Default sets default values for DSyncOptions.
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

// RedisClient sets the Redis client for DSyncOptions.
func (Option) RedisClient(cli *redis.Client) DSyncOption {
	return func(o *DSyncOptions) {
		o.RedisClient = cli
	}
}

// RedisConfig sets the Redis configuration options for DSyncOptions.
func (Option) RedisConfig(conf *redis.Options) DSyncOption {
	return func(o *DSyncOptions) {
		o.RedisConfig = conf
	}
}

// RedisURL sets the Redis server URL for DSyncOptions.
func (Option) RedisURL(url string) DSyncOption {
	return func(o *DSyncOptions) {
		o.RedisURL = url
	}
}

// KeyPrefix sets the key prefix for locking keys in DSyncOptions.
func (Option) KeyPrefix(prefix string) DSyncOption {
	return func(o *DSyncOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, ":") {
			prefix += ":"
		}
		o.KeyPrefix = prefix
	}
}

// FastAuth sets the username and password for authentication in DSyncOptions.
func (Option) FastAuth(username, password string) DSyncOption {
	return func(options *DSyncOptions) {
		options.FastUsername = username
		options.FastPassword = password
	}
}

// FastAddress sets the Redis server address in DSyncOptions.
func (Option) FastAddress(addr string) DSyncOption {
	return func(options *DSyncOptions) {
		if _, _, err := net.SplitHostPort(addr); err != nil {
			panic(err)
		}
		options.FastAddress = addr
	}
}

// FastDBIndex sets the Redis database index in DSyncOptions.
func (Option) FastDBIndex(idx int) DSyncOption {
	return func(options *DSyncOptions) {
		options.FastDBIndex = idx
	}
}
