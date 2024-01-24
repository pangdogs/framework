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
	RedisClient  *redis.Client
	RedisConfig  *redis.Options
	RedisURL     string
	KeyPrefix    string
	CustUsername string
	CustPassword string
	CustAddress  string
	CustDB       int
}

// Default sets default values for DSyncOptions.
func (Option) Default() option.Setting[DSyncOptions] {
	return func(options *DSyncOptions) {
		Option{}.RedisClient(nil)(options)
		Option{}.RedisConfig(nil)(options)
		Option{}.RedisURL("")(options)
		Option{}.KeyPrefix("golaxy:mutex:")(options)
		Option{}.CustAuth("", "")(options)
		Option{}.CustAddress("127.0.0.1:6379")(options)
		Option{}.CustDB(0)(options)
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

// CustAuth sets the username and password for authentication in DSyncOptions.
func (Option) CustAuth(username, password string) option.Setting[DSyncOptions] {
	return func(options *DSyncOptions) {
		options.CustUsername = username
		options.CustPassword = password
	}
}

// CustAddress sets the Redis server address in DSyncOptions.
func (Option) CustAddress(addr string) option.Setting[DSyncOptions] {
	return func(options *DSyncOptions) {
		if _, _, err := net.SplitHostPort(addr); err != nil {
			panic(err)
		}
		options.CustAddress = addr
	}
}

// CustDB sets the Redis database index in DSyncOptions.
func (Option) CustDB(db int) option.Setting[DSyncOptions] {
	return func(options *DSyncOptions) {
		options.CustDB = db
	}
}
