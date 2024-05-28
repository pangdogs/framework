package dent

import (
	"crypto/tls"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/option"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net"
	"strings"
	"time"
)

// DistEntitiesOptions 所有选项
type DistEntitiesOptions struct {
	EtcdClient      *clientv3.Client
	EtcdConfig      *clientv3.Config
	KeyPrefix       string
	TTL             time.Duration
	CustomUsername  string
	CustomPassword  string
	CustomAddresses []string
	CustomTLSConfig *tls.Config
}

var With _Option

type _Option struct{}

// Default 默认值
func (_Option) Default() option.Setting[DistEntitiesOptions] {
	return func(options *DistEntitiesOptions) {
		With.EtcdClient(nil)(options)
		With.EtcdConfig(nil)(options)
		With.KeyPrefix("/golaxy/entities/")(options)
		With.TTL(time.Minute)(options)
		With.CustomAuth("", "")(options)
		With.CustomAddresses("127.0.0.1:2379")(options)
		With.CustomTLSConfig(nil)(options)
	}
}

// EtcdClient etcd客户端，最优先使用
func (_Option) EtcdClient(cli *clientv3.Client) option.Setting[DistEntitiesOptions] {
	return func(o *DistEntitiesOptions) {
		o.EtcdClient = cli
	}
}

// EtcdConfig etcd配置，次优先使用
func (_Option) EtcdConfig(config *clientv3.Config) option.Setting[DistEntitiesOptions] {
	return func(o *DistEntitiesOptions) {
		o.EtcdConfig = config
	}
}

// KeyPrefix 所有key的前缀
func (_Option) KeyPrefix(prefix string) option.Setting[DistEntitiesOptions] {
	return func(options *DistEntitiesOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.KeyPrefix = prefix
	}
}

// TTL 实体信息过期时间
func (_Option) TTL(ttl time.Duration) option.Setting[DistEntitiesOptions] {
	return func(options *DistEntitiesOptions) {
		if ttl < 3*time.Second {
			panic(fmt.Errorf("%w: option TTL can't be set to a value less than 3 second", core.ErrArgs))
		}
		options.TTL = ttl
	}
}

// CustomAuth 自定义设置etcd鉴权信息
func (_Option) CustomAuth(username, password string) option.Setting[DistEntitiesOptions] {
	return func(options *DistEntitiesOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddresses 自定义设置etcd服务地址
func (_Option) CustomAddresses(addrs ...string) option.Setting[DistEntitiesOptions] {
	return func(options *DistEntitiesOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				panic(fmt.Errorf("%w: %w", core.ErrArgs, err))
			}
		}
		options.CustomAddresses = addrs
	}
}

// CustomTLSConfig 自定义设置加密etcd连接的配置
func (_Option) CustomTLSConfig(conf *tls.Config) option.Setting[DistEntitiesOptions] {
	return func(o *DistEntitiesOptions) {
		o.CustomTLSConfig = conf
	}
}
