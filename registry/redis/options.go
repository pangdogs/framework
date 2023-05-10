package etcd

import (
	"github.com/redis/go-redis/v9"
	"net"
	"strings"
	"time"
)

type RedisOptions struct {
	RedisClient   *redis.Client
	RedisConfig   *redis.Options
	RedisURL      string
	KeyPrefix     string
	Timeout       time.Duration
	WatchChanSize int
	FastUsername  string
	FastPassword  string
	FastAddress   string
	FastDBIndex   int
}

type RedisOption func(options *RedisOptions)

type WithRedisOption struct{}

func (WithRedisOption) Default() RedisOption {
	return func(options *RedisOptions) {
		WithRedisOption{}.RedisClient(nil)(options)
		WithRedisOption{}.RedisConfig(nil)(options)
		WithRedisOption{}.RedisURL("")(options)
		WithRedisOption{}.KeyPrefix("golaxy:registry:")(options)
		WithRedisOption{}.Timeout(3 * time.Second)(options)
		WithRedisOption{}.WatchChanSize(128)(options)
		WithRedisOption{}.FastAuth("", "")(options)
		WithRedisOption{}.FastAddress("127.0.0.1:6379")(options)
		WithRedisOption{}.FastDBIndex(0)(options)
	}
}

func (WithRedisOption) RedisClient(cli *redis.Client) RedisOption {
	return func(o *RedisOptions) {
		o.RedisClient = cli
	}
}

func (WithRedisOption) RedisConfig(conf *redis.Options) RedisOption {
	return func(o *RedisOptions) {
		o.RedisConfig = conf
	}
}

func (WithRedisOption) RedisURL(url string) RedisOption {
	return func(o *RedisOptions) {
		o.RedisURL = url
	}
}

func (WithRedisOption) KeyPrefix(prefix string) RedisOption {
	return func(o *RedisOptions) {
		if !strings.HasSuffix(prefix, ":") {
			prefix += ":"
		}
		o.KeyPrefix = prefix
	}
}

func (WithRedisOption) Timeout(dur time.Duration) RedisOption {
	return func(o *RedisOptions) {
		o.Timeout = dur
	}
}

func (WithRedisOption) WatchChanSize(size int) RedisOption {
	return func(o *RedisOptions) {
		if size < 0 {
			panic("options.WatchChanSize can't be set to a value less then 0")
		}
		o.WatchChanSize = size
	}
}

func (WithRedisOption) FastAuth(username, password string) RedisOption {
	return func(options *RedisOptions) {
		options.FastUsername = username
		options.FastPassword = password
	}
}

func (WithRedisOption) FastAddress(addr string) RedisOption {
	return func(options *RedisOptions) {
		if _, _, err := net.SplitHostPort(addr); err != nil {
			panic(err)
		}
		options.FastAddress = addr
	}
}

func (WithRedisOption) FastDBIndex(idx int) RedisOption {
	return func(options *RedisOptions) {
		options.FastDBIndex = idx
	}
}
