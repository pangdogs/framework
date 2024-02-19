package etcd_dsync

import (
	"crypto/tls"
	"git.golaxy.org/core/util/option"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net"
	"strings"
)

// DSyncOptions contains various options for configuring distributed locking using etcd.
type DSyncOptions struct {
	EtcdClient      *clientv3.Client
	EtcdConfig      *clientv3.Config
	KeyPrefix       string
	WatchChanSize   int
	CustomUsername  string
	CustomPassword  string
	CustomAddresses []string
	CustomSecure    bool
	CustomTLSConfig *tls.Config
}

var With _Option

type _Option struct{}

// Default sets default values for DSyncOptions.
func (_Option) Default() option.Setting[DSyncOptions] {
	return func(options *DSyncOptions) {
		With.EtcdClient(nil)(options)
		With.EtcdConfig(nil)(options)
		With.KeyPrefix("/golaxy/mutex/")(options)
		With.CustomAuth("", "")(options)
		With.CustomAddresses("127.0.0.1:2379")(options)
		With.CustomSecure(false)(options)
		With.CustomTLSConfig(nil)(options)
	}
}

// EtcdClient sets the etcd client for DSyncOptions.
func (_Option) EtcdClient(cli *clientv3.Client) option.Setting[DSyncOptions] {
	return func(o *DSyncOptions) {
		o.EtcdClient = cli
	}
}

// EtcdConfig sets the etcd config for DSyncOptions.
func (_Option) EtcdConfig(config *clientv3.Config) option.Setting[DSyncOptions] {
	return func(o *DSyncOptions) {
		o.EtcdConfig = config
	}
}

// KeyPrefix sets the key prefix for locking keys in DSyncOptions.
func (_Option) KeyPrefix(prefix string) option.Setting[DSyncOptions] {
	return func(options *DSyncOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.KeyPrefix = prefix
	}
}

// CustomAuth sets the username and password for authentication in DSyncOptions.
func (_Option) CustomAuth(username, password string) option.Setting[DSyncOptions] {
	return func(options *DSyncOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddresses sets the etcd server addresses in DSyncOptions.
func (_Option) CustomAddresses(addrs ...string) option.Setting[DSyncOptions] {
	return func(options *DSyncOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				panic(err)
			}
		}
		options.CustomAddresses = addrs
	}
}

// CustomSecure sets whether to use a secure connection (HTTPS) in DSyncOptions.
func (_Option) CustomSecure(secure bool) option.Setting[DSyncOptions] {
	return func(o *DSyncOptions) {
		o.CustomSecure = secure
	}
}

// CustomTLSConfig sets the TLS configuration for secure connections in DSyncOptions.
func (_Option) CustomTLSConfig(conf *tls.Config) option.Setting[DSyncOptions] {
	return func(o *DSyncOptions) {
		o.CustomTLSConfig = conf
	}
}
