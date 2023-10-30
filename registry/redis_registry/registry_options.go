package redis_registry

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"kit.golaxy.org/golaxy"
	"net"
	"strings"
)

// Option 所有选项设置器
type Option struct{}

// RegistryOptions 所有选项
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

// RegistryOption 选项设置器
type RegistryOption func(options *RegistryOptions)

// Default 默认值
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

// RedisClient redis客户端，1st优先使用
func (Option) RedisClient(cli *redis.Client) RegistryOption {
	return func(o *RegistryOptions) {
		o.RedisClient = cli
	}
}

// RedisConfig redis配置，2nd优先使用
func (Option) RedisConfig(conf *redis.Options) RegistryOption {
	return func(o *RegistryOptions) {
		o.RedisConfig = conf
	}
}

// RedisURL redis连接url，3rd优先使用
func (Option) RedisURL(url string) RegistryOption {
	return func(o *RegistryOptions) {
		o.RedisURL = url
	}
}

// KeyPrefix 所有key的前缀
func (Option) KeyPrefix(prefix string) RegistryOption {
	return func(o *RegistryOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, ":") {
			prefix += ":"
		}
		o.KeyPrefix = prefix
	}
}

// WatchChanSize 监控服务变化的channel大小
func (Option) WatchChanSize(size int) RegistryOption {
	return func(o *RegistryOptions) {
		if size < 0 {
			panic(fmt.Errorf("%w: option WatchChanSize can't be set to a value less then 0", golaxy.ErrArgs))
		}
		o.WatchChanSize = size
	}
}

// FastAuth 快速设置redis鉴权信息
func (Option) FastAuth(username, password string) RegistryOption {
	return func(options *RegistryOptions) {
		options.FastUsername = username
		options.FastPassword = password
	}
}

// FastAddress 快速设置redis服务地址
func (Option) FastAddress(addr string) RegistryOption {
	return func(options *RegistryOptions) {
		if _, _, err := net.SplitHostPort(addr); err != nil {
			panic(fmt.Errorf("%w: %w", golaxy.ErrArgs, err))
		}
		options.FastAddress = addr
	}
}

// FastDBIndex 快速设置redis db下标
func (Option) FastDBIndex(idx int) RegistryOption {
	return func(options *RegistryOptions) {
		options.FastDBIndex = idx
	}
}
