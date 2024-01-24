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
	CustUsername  string
	CustPassword  string
	CustAddresses []string
	CustSecure    bool
	CustTLSConfig *tls.Config
}

// Default sets default values for DSyncOptions.
func (Option) Default() option.Setting[DSyncOptions] {
	return func(options *DSyncOptions) {
		Option{}.EtcdClient(nil)(options)
		Option{}.EtcdConfig(nil)(options)
		Option{}.KeyPrefix("/golaxy/mutex/")(options)
		Option{}.CustAuth("", "")(options)
		Option{}.CustAddresses("127.0.0.1:2379")(options)
		Option{}.CustSecure(false)(options)
		Option{}.CustTLSConfig(nil)(options)
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

// CustAuth sets the username and password for authentication in DSyncOptions.
func (Option) CustAuth(username, password string) option.Setting[DSyncOptions] {
	return func(options *DSyncOptions) {
		options.CustUsername = username
		options.CustPassword = password
	}
}

// CustAddresses sets the etcd server addresses in DSyncOptions.
func (Option) CustAddresses(addrs ...string) option.Setting[DSyncOptions] {
	return func(options *DSyncOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				panic(err)
			}
		}
		options.CustAddresses = addrs
	}
}

// CustSecure sets whether to use a secure connection (HTTPS) in DSyncOptions.
func (Option) CustSecure(secure bool) option.Setting[DSyncOptions] {
	return func(o *DSyncOptions) {
		o.CustSecure = secure
	}
}

// CustTLSConfig sets the TLS configuration for secure connections in DSyncOptions.
func (Option) CustTLSConfig(conf *tls.Config) option.Setting[DSyncOptions] {
	return func(o *DSyncOptions) {
		o.CustTLSConfig = conf
	}
}
