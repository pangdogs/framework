package dentq

import (
	"crypto/tls"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/util/option"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net"
	"strings"
	"time"
)

// Option 所有选项设置器
type Option struct{}

// DistEntityQuerierOptions 所有选项
type DistEntityQuerierOptions struct {
	EtcdClient      *clientv3.Client
	EtcdConfig      *clientv3.Config
	KeyPrefix       string
	CacheExpiry     time.Duration
	CustomUsername  string
	CustomPassword  string
	CustomAddresses []string
	CustomSecure    bool
	CustomTLSConfig *tls.Config
}

// Default 默认值
func (Option) Default() option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		Option{}.EtcdClient(nil)(options)
		Option{}.EtcdConfig(nil)(options)
		Option{}.KeyPrefix("/golaxy/entities/")(options)
		Option{}.CacheExpiry(10 * time.Minute)(options)
		Option{}.CustomAuth("", "")(options)
		Option{}.CustomAddresses("127.0.0.1:2379")(options)
		Option{}.CustomSecure(false)(options)
		Option{}.CustomTLSConfig(nil)(options)
	}
}

// EtcdClient etcd客户端，最优先使用
func (Option) EtcdClient(cli *clientv3.Client) option.Setting[DistEntityQuerierOptions] {
	return func(o *DistEntityQuerierOptions) {
		o.EtcdClient = cli
	}
}

// EtcdConfig etcd配置，次优先使用
func (Option) EtcdConfig(config *clientv3.Config) option.Setting[DistEntityQuerierOptions] {
	return func(o *DistEntityQuerierOptions) {
		o.EtcdConfig = config
	}
}

// KeyPrefix 所有key的前缀
func (Option) KeyPrefix(prefix string) option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.KeyPrefix = prefix
	}
}

// CacheExpiry 缓存过期时间
func (Option) CacheExpiry(expiry time.Duration) option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		options.CacheExpiry = expiry
	}
}

// CustomAuth 自定义设置etcd鉴权信息
func (Option) CustomAuth(username, password string) option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddresses 自定义设置etcd服务地址
func (Option) CustomAddresses(addrs ...string) option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				panic(fmt.Errorf("%w: %w", core.ErrArgs, err))
			}
		}
		options.CustomAddresses = addrs
	}
}

// CustomSecure 自定义设置是否加密etcd连接
func (Option) CustomSecure(secure bool) option.Setting[DistEntityQuerierOptions] {
	return func(o *DistEntityQuerierOptions) {
		o.CustomSecure = secure
	}
}

// CustomTLSConfig 自定义设置加密etcd连接的配置
func (Option) CustomTLSConfig(conf *tls.Config) option.Setting[DistEntityQuerierOptions] {
	return func(o *DistEntityQuerierOptions) {
		o.CustomTLSConfig = conf
	}
}