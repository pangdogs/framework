package router

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

type RouterOptions struct {
	EtcdClient          *clientv3.Client
	EtcdConfig          *clientv3.Config
	KeyPrefix           string
	WatchChanSize       int
	GroupTTL            time.Duration
	GroupAutoRefreshTTL bool
	CustomUsername      string
	CustomPassword      string
	CustomAddresses     []string
	CustomTLSConfig     *tls.Config
}

var With _Option

type _Option struct{}

// Default 默认值
func (_Option) Default() option.Setting[RouterOptions] {
	return func(options *RouterOptions) {
		With.EtcdClient(nil)(options)
		With.EtcdConfig(nil)(options)
		With.KeyPrefix("/golaxy/groups/")(options)
		With.WatchChanSize(128)(options)
		With.GroupTTL(30*time.Second, true)(options)
		With.CustomAuth("", "")(options)
		With.CustomAddresses("127.0.0.1:2379")(options)
		With.CustomTLSConfig(nil)(options)
	}
}

// EtcdClient etcd客户端，最优先使用
func (_Option) EtcdClient(cli *clientv3.Client) option.Setting[RouterOptions] {
	return func(o *RouterOptions) {
		o.EtcdClient = cli
	}
}

// EtcdConfig etcd配置，次优先使用
func (_Option) EtcdConfig(config *clientv3.Config) option.Setting[RouterOptions] {
	return func(o *RouterOptions) {
		o.EtcdConfig = config
	}
}

// KeyPrefix 所有key的前缀
func (_Option) KeyPrefix(prefix string) option.Setting[RouterOptions] {
	return func(options *RouterOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.KeyPrefix = prefix
	}
}

// WatchChanSize 监控服务变化的channel大小
func (_Option) WatchChanSize(size int) option.Setting[RouterOptions] {
	return func(options *RouterOptions) {
		if size < 0 {
			panic(fmt.Errorf("%w: option WatchChanSize can't be set to a value less than 0", core.ErrArgs))
		}
		options.WatchChanSize = size
	}
}

// GroupTTL 分组默认TTL
func (_Option) GroupTTL(ttl time.Duration, auto bool) option.Setting[RouterOptions] {
	return func(options *RouterOptions) {
		if ttl < 3*time.Second {
			panic(fmt.Errorf("%w: option GroupTTL can't be set to a value less than 3 second", core.ErrArgs))
		}
		options.GroupTTL = ttl
		options.GroupAutoRefreshTTL = auto
	}
}

// CustomAuth 自定义设置etcd鉴权信息
func (_Option) CustomAuth(username, password string) option.Setting[RouterOptions] {
	return func(options *RouterOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddresses 自定义设置etcd服务地址
func (_Option) CustomAddresses(addrs ...string) option.Setting[RouterOptions] {
	return func(options *RouterOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				panic(fmt.Errorf("%w: %w", core.ErrArgs, err))
			}
		}
		options.CustomAddresses = addrs
	}
}

// CustomTLSConfig 自定义设置加密etcd连接的配置
func (_Option) CustomTLSConfig(conf *tls.Config) option.Setting[RouterOptions] {
	return func(o *RouterOptions) {
		o.CustomTLSConfig = conf
	}
}
