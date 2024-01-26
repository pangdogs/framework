package dent

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

// DistEntitiesOptions 所有选项
type DistEntitiesOptions struct {
	EtcdClient      *clientv3.Client
	EtcdConfig      *clientv3.Config
	KeyPrefix       string
	TTL             time.Duration
	CustomUsername  string
	CustomPassword  string
	CustomAddresses []string
	CustomSecure    bool
	CustomTLSConfig *tls.Config
}

// Default 默认值
func (Option) Default() option.Setting[DistEntitiesOptions] {
	return func(options *DistEntitiesOptions) {
		Option{}.EtcdClient(nil)(options)
		Option{}.EtcdConfig(nil)(options)
		Option{}.KeyPrefix("/golaxy/entities/")(options)
		Option{}.TTL(time.Minute)(options)
		Option{}.CustomAuth("", "")(options)
		Option{}.CustomAddresses("127.0.0.1:2379")(options)
		Option{}.CustomSecure(false)(options)
		Option{}.CustomTLSConfig(nil)(options)
	}
}

// EtcdClient etcd客户端，最优先使用
func (Option) EtcdClient(cli *clientv3.Client) option.Setting[DistEntitiesOptions] {
	return func(o *DistEntitiesOptions) {
		o.EtcdClient = cli
	}
}

// EtcdConfig etcd配置，次优先使用
func (Option) EtcdConfig(config *clientv3.Config) option.Setting[DistEntitiesOptions] {
	return func(o *DistEntitiesOptions) {
		o.EtcdConfig = config
	}
}

// KeyPrefix 所有key的前缀
func (Option) KeyPrefix(prefix string) option.Setting[DistEntitiesOptions] {
	return func(options *DistEntitiesOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.KeyPrefix = prefix
	}
}

// TTL 实体信息过期时间
func (Option) TTL(ttl time.Duration) option.Setting[DistEntitiesOptions] {
	return func(options *DistEntitiesOptions) {
		if ttl < 3*time.Second {
			panic(fmt.Errorf("%w: option TTL can't be set to a value less than 3 second", core.ErrArgs))
		}
		options.TTL = ttl
	}
}

// CustomAuth 自定义设置etcd鉴权信息
func (Option) CustomAuth(username, password string) option.Setting[DistEntitiesOptions] {
	return func(options *DistEntitiesOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddresses 自定义设置etcd服务地址
func (Option) CustomAddresses(addrs ...string) option.Setting[DistEntitiesOptions] {
	return func(options *DistEntitiesOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				panic(fmt.Errorf("%w: %w", core.ErrArgs, err))
			}
		}
		options.CustomAddresses = addrs
	}
}

// CustomSecure 自定义设置是否加密etcd连接
func (Option) CustomSecure(secure bool) option.Setting[DistEntitiesOptions] {
	return func(o *DistEntitiesOptions) {
		o.CustomSecure = secure
	}
}

// CustomTLSConfig 自定义设置加密etcd连接的配置
func (Option) CustomTLSConfig(conf *tls.Config) option.Setting[DistEntitiesOptions] {
	return func(o *DistEntitiesOptions) {
		o.CustomTLSConfig = conf
	}
}
