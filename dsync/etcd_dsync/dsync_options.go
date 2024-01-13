package etcd_dsync

import (
	"crypto/tls"
	"git.golaxy.org/core/util/option"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net"
	"strings"
)

// Option is a struct used for setting options.
type Option struct{}

// DSyncOptions contains various options for configuring distributed locking using etcd.
type DSyncOptions struct {
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

// Default sets default values for DSyncOptions.
func (Option) Default() option.Setting[DSyncOptions] {
	return func(options *DSyncOptions) {
		Option{}.EtcdClient(nil)(options)
		Option{}.EtcdConfig(nil)(options)
		Option{}.KeyPrefix("/golaxy/mutex/")(options)
		Option{}.FastAuth("", "")(options)
		Option{}.FastAddresses("127.0.0.1:2379")(options)
		Option{}.FastSecure(false)(options)
		Option{}.FastTLSConfig(nil)(options)
	}
}

// EtcdClient sets the etcd client for DSyncOptions.
func (Option) EtcdClient(cli *clientv3.Client) option.Setting[DSyncOptions] {
	return func(o *DSyncOptions) {
		o.EtcdClient = cli
	}
}

// EtcdConfig sets the etcd config for DSyncOptions.
func (Option) EtcdConfig(config *clientv3.Config) option.Setting[DSyncOptions] {
	return func(o *DSyncOptions) {
		o.EtcdConfig = config
	}
}

// KeyPrefix sets the key prefix for locking keys in DSyncOptions.
func (Option) KeyPrefix(prefix string) option.Setting[DSyncOptions] {
	return func(options *DSyncOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.KeyPrefix = prefix
	}
}

// FastAuth sets the username and password for authentication in DSyncOptions.
func (Option) FastAuth(username, password string) option.Setting[DSyncOptions] {
	return func(options *DSyncOptions) {
		options.FastUsername = username
		options.FastPassword = password
	}
}

// FastAddresses sets the etcd server addresses in DSyncOptions.
func (Option) FastAddresses(addrs ...string) option.Setting[DSyncOptions] {
	return func(options *DSyncOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				panic(err)
			}
		}
		options.FastAddresses = addrs
	}
}

// FastSecure sets whether to use a secure connection (HTTPS) in DSyncOptions.
func (Option) FastSecure(secure bool) option.Setting[DSyncOptions] {
	return func(o *DSyncOptions) {
		o.FastSecure = secure
	}
}

// FastTLSConfig sets the TLS configuration for secure connections in DSyncOptions.
func (Option) FastTLSConfig(conf *tls.Config) option.Setting[DSyncOptions] {
	return func(o *DSyncOptions) {
		o.FastTLSConfig = conf
	}
}
