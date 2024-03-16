package redis_discovery

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/util/option"
	"github.com/redis/go-redis/v9"
	"net"
	"strings"
	"time"
)

// RegistryOptions 所有选项
type RegistryOptions struct {
	RedisClient    *redis.Client
	RedisConfig    *redis.Options
	RedisURL       string
	KeyPrefix      string
	WatchChanSize  int
	TTL            time.Duration
	CustomUsername string
	CustomPassword string
	CustomAddress  string
	CustomDB       int
}

var With _Option

type _Option struct{}

// Default 默认值
func (_Option) Default() option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		With.RedisClient(nil)(options)
		With.RedisConfig(nil)(options)
		With.RedisURL("")(options)
		With.KeyPrefix("golaxy:services:")(options)
		With.WatchChanSize(128)(options)
		With.TTL(10 * time.Second)(options)
		With.CustomAuth("", "")(options)
		With.CustomAddress("127.0.0.1:6379")(options)
		With.CustomDB(0)(options)
	}
}

// RedisClient redis客户端，1st优先使用
func (_Option) RedisClient(cli *redis.Client) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		o.RedisClient = cli
	}
}

// RedisConfig redis配置，2nd优先使用
func (_Option) RedisConfig(conf *redis.Options) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		o.RedisConfig = conf
	}
}

// RedisURL redis连接url，3rd优先使用
func (_Option) RedisURL(url string) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		o.RedisURL = url
	}
}

// KeyPrefix 所有key的前缀
func (_Option) KeyPrefix(prefix string) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, ":") {
			prefix += ":"
		}
		o.KeyPrefix = prefix
	}
}

// WatchChanSize 监控服务变化的channel大小
func (_Option) WatchChanSize(size int) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		if size < 0 {
			panic(fmt.Errorf("%w: option WatchChanSize can't be set to a value less than 0", core.ErrArgs))
		}
		o.WatchChanSize = size
	}
}

// TTL 默认TTL
func (_Option) TTL(ttl time.Duration) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		if ttl < 3*time.Second {
			panic(fmt.Errorf("%w: option TTL can't be set to a value less than 3 second", core.ErrArgs))
		}
		options.TTL = ttl
	}
}

// CustomAuth 自定义设置redis鉴权信息
func (_Option) CustomAuth(username, password string) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddress 自定义设置redis服务地址
func (_Option) CustomAddress(addr string) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		if _, _, err := net.SplitHostPort(addr); err != nil {
			panic(fmt.Errorf("%w: %w", core.ErrArgs, err))
		}
		options.CustomAddress = addr
	}
}

// CustomDB 自定义设置redis db
func (_Option) CustomDB(db int) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		options.CustomDB = db
	}
}
