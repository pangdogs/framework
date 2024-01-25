package etcd_discovery

import (
	"crypto/tls"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/util/option"
	clientv3 "go.etcd.io/etcd/client/v3"
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
	CustUsername  string
	CustPassword  string
	CustAddresses []string
	CustSecure    bool
	CustTLSConfig *tls.Config
}

// Default 默认值
func (Option) Default() option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		Option{}.EtcdClient(nil)(options)
		Option{}.EtcdConfig(nil)(options)
		Option{}.KeyPrefix("/golaxy/registry/")(options)
		Option{}.WatchChanSize(128)(options)
		Option{}.CustAuth("", "")(options)
		Option{}.CustAddresses("127.0.0.1:2379")(options)
		Option{}.CustSecure(false)(options)
		Option{}.CustTLSConfig(nil)(options)
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
			panic(fmt.Errorf("%w: option WatchChanSize can't be set to a value less than 0", core.ErrArgs))
		}
		options.WatchChanSize = size
	}
}

// CustAuth 自定义设置etcd鉴权信息
func (Option) CustAuth(username, password string) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		options.CustUsername = username
		options.CustPassword = password
	}
}

// CustAddresses 自定义设置etcd服务地址
func (Option) CustAddresses(addrs ...string) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				panic(fmt.Errorf("%w: %w", core.ErrArgs, err))
			}
		}
		options.CustAddresses = addrs
	}
}

// CustSecure 自定义设置是否加密etcd连接
func (Option) CustSecure(secure bool) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		o.CustSecure = secure
	}
}

// CustTLSConfig 自定义设置加密etcd连接的配置
func (Option) CustTLSConfig(conf *tls.Config) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		o.CustTLSConfig = conf
	}
}
