package etcd

import (
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

type RedisOptions struct {
	RedisClient *redis.Client
	RedisConfig *redis.Options
	RedisURL    string
	KeyPrefix   string
	Timeout     time.Duration
}

type RedisOption func(options *RedisOptions)

type WithRedisOption struct{}

func (WithRedisOption) Default() RedisOption {
	return func(options *RedisOptions) {
		WithRedisOption{}.RedisClient(nil)(options)
		WithRedisOption{}.RedisConfig(nil)(options)
		WithRedisOption{}.RedisURL("")(options)
		WithRedisOption{}.KeyPrefix("golaxy:registry:")
		WithRedisOption{}.Timeout(3 * time.Second)(options)
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
		if dur <= 0 {
			panic("options.Timeout can't be set to a value less equal 0")
		}
		o.Timeout = dur
	}
}
