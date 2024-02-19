package redis_dsync

import (
	"git.golaxy.org/core/util/option"
	"github.com/redis/go-redis/v9"
	"net"
	"strings"
)

// DSyncOptions contains various options for configuring distributed locking using redis.
type DSyncOptions struct {
	RedisClient    *redis.Client
	RedisConfig    *redis.Options
	RedisURL       string
	KeyPrefix      string
	CustomUsername string
	CustomPassword string
	CustomAddress  string
	CustomDB       int
}

var With _Option

type _Option struct{}

// Default sets default values for DSyncOptions.
func (_Option) Default() option.Setting[DSyncOptions] {
	return func(options *DSyncOptions) {
		With.RedisClient(nil)(options)
		With.RedisConfig(nil)(options)
		With.RedisURL("")(options)
		With.KeyPrefix("golaxy:mutex:")(options)
		With.CustomAuth("", "")(options)
		With.CustomAddress("127.0.0.1:6379")(options)
		With.CustomDB(0)(options)
	}
}

// RedisClient sets the Redis client for DSyncOptions.
func (_Option) RedisClient(cli *redis.Client) option.Setting[DSyncOptions] {
	return func(o *DSyncOptions) {
		o.RedisClient = cli
	}
}

// RedisConfig sets the Redis configuration options for DSyncOptions.
func (_Option) RedisConfig(conf *redis.Options) option.Setting[DSyncOptions] {
	return func(o *DSyncOptions) {
		o.RedisConfig = conf
	}
}

// RedisURL sets the Redis server URL for DSyncOptions.
func (_Option) RedisURL(url string) option.Setting[DSyncOptions] {
	return func(o *DSyncOptions) {
		o.RedisURL = url
	}
}

// KeyPrefix sets the key prefix for locking keys in DSyncOptions.
func (_Option) KeyPrefix(prefix string) option.Setting[DSyncOptions] {
	return func(o *DSyncOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, ":") {
			prefix += ":"
		}
		o.KeyPrefix = prefix
	}
}

// CustomAuth sets the username and password for authentication in DSyncOptions.
func (_Option) CustomAuth(username, password string) option.Setting[DSyncOptions] {
	return func(options *DSyncOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddress sets the Redis server address in DSyncOptions.
func (_Option) CustomAddress(addr string) option.Setting[DSyncOptions] {
	return func(options *DSyncOptions) {
		if _, _, err := net.SplitHostPort(addr); err != nil {
			panic(err)
		}
		options.CustomAddress = addr
	}
}

// CustomDB sets the Redis database index in DSyncOptions.
func (_Option) CustomDB(db int) option.Setting[DSyncOptions] {
	return func(options *DSyncOptions) {
		options.CustomDB = db
	}
}
