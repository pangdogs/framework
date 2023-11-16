package etcd_registry

import (
	"crypto/tls"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"kit.golaxy.org/golaxy"
	"kit.golaxy.org/golaxy/util/option"
	"net"
	"strings"
)

// Option 所有选项设置器
type Option struct{}

// RegistryOptions 所有选项
type RegistryOptions struct {
	EtcdClient    *clientv3.Client
	EtcdConfig    *clientv3.Config
	KeyPrefix     string
	WatchChanSize int
	FastUsername  string
	FastPassword  string
	FastAddresses []string
	FastSecure    bool
	FastTLSConfig *tls.Config
}

// Default 默认值
func (Option) Default() option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		Option{}.EtcdClient(nil)(options)
		Option{}.EtcdConfig(nil)(options)
		Option{}.KeyPrefix("/golaxy/registry/")(options)
		Option{}.WatchChanSize(128)(options)
		Option{}.FastAuth("", "")(options)
		Option{}.FastAddresses("127.0.0.1:2379")(options)
		Option{}.FastSecure(false)(options)
		Option{}.FastTLSConfig(nil)(options)
	}
}

// EtcdClient etcd客户端，最优先使用
func (Option) EtcdClient(cli *clientv3.Client) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		o.EtcdClient = cli
	}
}

// EtcdConfig etcd配置，次优先使用
func (Option) EtcdConfig(config *clientv3.Config) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		o.EtcdConfig = config
	}
}

// KeyPrefix 所有key的前缀
func (Option) KeyPrefix(prefix string) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.KeyPrefix = prefix
	}
}

// WatchChanSize 监控服务变化的channel大小
func (Option) WatchChanSize(size int) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		if size < 0 {
			panic(fmt.Errorf("%w: option WatchChanSize can't be set to a value less then 0", golaxy.ErrArgs))
		}
		options.WatchChanSize = size
	}
}

// FastAuth 快速设置etcd鉴权信息
func (Option) FastAuth(username, password string) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		options.FastUsername = username
		options.FastPassword = password
	}
}

// FastAddresses 快速设置etcd服务地址
func (Option) FastAddresses(addrs ...string) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				panic(fmt.Errorf("%w: %w", golaxy.ErrArgs, err))
			}
		}
		options.FastAddresses = addrs
	}
}

// FastSecure 快速设置是否加密etcd连接
func (Option) FastSecure(secure bool) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		o.FastSecure = secure
	}
}

// FastTLSConfig 快速设置加密etcd连接的配置
func (Option) FastTLSConfig(conf *tls.Config) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		o.FastTLSConfig = conf
	}
}
