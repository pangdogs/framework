package redis_dsync

import (
	"git.golaxy.org/core/util/option"
	"github.com/redis/go-redis/v9"
	"net"
	"strings"
)

// Option is a struct used for setting options.
type Option struct{}

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

// Default sets default values for DSyncOptions.
func (Option) Default() option.Setting[DSyncOptions] {
	return func(options *DSyncOptions) {
		Option{}.RedisClient(nil)(options)
		Option{}.RedisConfig(nil)(options)
		Option{}.RedisURL("")(options)
		Option{}.KeyPrefix("golaxy:mutex:")(options)
		Option{}.CustomAuth("", "")(options)
		Option{}.CustomAddress("127.0.0.1:6379")(options)
		Option{}.CustomDB(0)(options)
	}
}

// RedisClient sets the Redis client for DSyncOptions.
func (Option) RedisClient(cli *redis.Client) option.Setting[DSyncOptions] {
	return func(o *DSyncOptions) {
		o.RedisClient = cli
	}
}

// RedisConfig sets the Redis configuration options for DSyncOptions.
func (Option) RedisConfig(conf *redis.Options) option.Setting[DSyncOptions] {
	return func(o *DSyncOptions) {
		o.RedisConfig = conf
	}
}

// RedisURL sets the Redis server URL for DSyncOptions.
func (Option) RedisURL(url string) option.Setting[DSyncOptions] {
	return func(o *DSyncOptions) {
		o.RedisURL = url
	}
}

// KeyPrefix sets the key prefix for locking keys in DSyncOptions.
func (Option) KeyPrefix(prefix string) option.Setting[DSyncOptions] {
	return func(o *DSyncOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, ":") {
			prefix += ":"
		}
		o.KeyPrefix = prefix
	}
}

// CustomAuth sets the username and password for authentication in DSyncOptions.
func (Option) CustomAuth(username, password string) option.Setting[DSyncOptions] {
	return func(options *DSyncOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddress sets the Redis server address in DSyncOptions.
func (Option) CustomAddress(addr string) option.Setting[DSyncOptions] {
	return func(options *DSyncOptions) {
		if _, _, err := net.SplitHostPort(addr); err != nil {
			panic(err)
		}
		options.CustomAddress = addr
	}
}

// CustomDB sets the Redis database index in DSyncOptions.
func (Option) CustomDB(db int) option.Setting[DSyncOptions] {
	return func(options *DSyncOptions) {
		options.CustomDB = db
	}
}
