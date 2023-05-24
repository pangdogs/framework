package etcd

import (
	"crypto/tls"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net"
	"strings"
)

type WithOption struct{}

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

type RegistryOption func(options *RegistryOptions)

func (WithOption) Default() RegistryOption {
	return func(options *RegistryOptions) {
		WithOption{}.EtcdClient(nil)(options)
		WithOption{}.EtcdConfig(nil)(options)
		WithOption{}.KeyPrefix("/golaxy/registry/")(options)
		WithOption{}.WatchChanSize(128)(options)
		WithOption{}.FastAuth("", "")(options)
		WithOption{}.FastAddresses("127.0.0.1:2379")(options)
		WithOption{}.FastSecure(false)(options)
		WithOption{}.FastTLSConfig(nil)(options)
	}
}

func (WithOption) EtcdClient(cli *clientv3.Client) RegistryOption {
	return func(o *RegistryOptions) {
		o.EtcdClient = cli
	}
}

func (WithOption) EtcdConfig(config *clientv3.Config) RegistryOption {
	return func(o *RegistryOptions) {
		o.EtcdConfig = config
	}
}

func (WithOption) KeyPrefix(prefix string) RegistryOption {
	return func(options *RegistryOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.KeyPrefix = prefix
	}
}

func (WithOption) WatchChanSize(size int) RegistryOption {
	return func(options *RegistryOptions) {
		if size < 0 {
			panic("option WatchChanSize can't be set to a value less then 0")
		}
		options.WatchChanSize = size
	}
}

func (WithOption) FastAuth(username, password string) RegistryOption {
	return func(options *RegistryOptions) {
		options.FastUsername = username
		options.FastPassword = password
	}
}

func (WithOption) FastAddresses(addrs ...string) RegistryOption {
	return func(options *RegistryOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				panic(err)
			}
		}
		options.FastAddresses = addrs
	}
}

func (WithOption) FastSecure(secure bool) RegistryOption {
	return func(o *RegistryOptions) {
		o.FastSecure = secure
	}
}

func (WithOption) FastTLSConfig(conf *tls.Config) RegistryOption {
	return func(o *RegistryOptions) {
		o.FastTLSConfig = conf
	}
}
