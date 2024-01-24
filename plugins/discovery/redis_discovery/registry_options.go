package redis_discovery

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/util/option"
	"github.com/redis/go-redis/v9"
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
	CustUsername  string
	CustPassword  string
	CustAddress   string
	CustDB        int
}

// Default 默认值
func (Option) Default() option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		Option{}.RedisClient(nil)(options)
		Option{}.RedisConfig(nil)(options)
		Option{}.RedisURL("")(options)
		Option{}.KeyPrefix("golaxy:registry:")(options)
		Option{}.WatchChanSize(128)(options)
		Option{}.CustAuth("", "")(options)
		Option{}.CustAddress("127.0.0.1:6379")(options)
		Option{}.CustDB(0)(options)
	}
}

// RedisClient redis客户端，1st优先使用
func (Option) RedisClient(cli *redis.Client) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		o.RedisClient = cli
	}
}

// RedisConfig redis配置，2nd优先使用
func (Option) RedisConfig(conf *redis.Options) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		o.RedisConfig = conf
	}
}

// RedisURL redis连接url，3rd优先使用
func (Option) RedisURL(url string) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		o.RedisURL = url
	}
}

// KeyPrefix 所有key的前缀
func (Option) KeyPrefix(prefix string) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, ":") {
			prefix += ":"
		}
		o.KeyPrefix = prefix
	}
}

// WatchChanSize 监控服务变化的channel大小
func (Option) WatchChanSize(size int) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		if size < 0 {
			panic(fmt.Errorf("%w: option WatchChanSize can't be set to a value less then 0", core.ErrArgs))
		}
		o.WatchChanSize = size
	}
}

// CustAuth 自定义设置redis鉴权信息
func (Option) CustAuth(username, password string) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		options.CustUsername = username
		options.CustPassword = password
	}
}

// CustAddress 自定义设置redis服务地址
func (Option) CustAddress(addr string) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		if _, _, err := net.SplitHostPort(addr); err != nil {
			panic(fmt.Errorf("%w: %w", core.ErrArgs, err))
		}
		options.CustAddress = addr
	}
}

// CustDB 自定义设置redis db
func (Option) CustDB(db int) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		options.CustDB = db
	}
}
