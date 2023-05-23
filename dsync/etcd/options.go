package etcd

import (
	"crypto/tls"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net"
	"strings"
)

type WithOption struct{}

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

type DSyncOption func(options *DSyncOptions)

func (WithOption) Default() DSyncOption {
	return func(options *DSyncOptions) {
		WithOption{}.EtcdClient(nil)(options)
		WithOption{}.EtcdConfig(nil)(options)
		WithOption{}.KeyPrefix("/golaxy/mutex/")(options)
		WithOption{}.WatchChanSize(128)(options)
		WithOption{}.FastAuth("", "")(options)
		WithOption{}.FastAddresses("127.0.0.1:2379")(options)
		WithOption{}.FastSecure(false)(options)
		WithOption{}.FastTLSConfig(nil)(options)
	}
}

func (WithOption) EtcdClient(cli *clientv3.Client) DSyncOption {
	return func(o *DSyncOptions) {
		o.EtcdClient = cli
	}
}

func (WithOption) EtcdConfig(config *clientv3.Config) DSyncOption {
	return func(o *DSyncOptions) {
		o.EtcdConfig = config
	}
}

func (WithOption) KeyPrefix(prefix string) DSyncOption {
	return func(options *DSyncOptions) {
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.KeyPrefix = prefix
	}
}

func (WithOption) WatchChanSize(size int) DSyncOption {
	return func(options *DSyncOptions) {
		if size < 0 {
			panic("option WatchChanSize can't be set to a value less then 0")
		}
		options.WatchChanSize = size
	}
}

func (WithOption) FastAuth(username, password string) DSyncOption {
	return func(options *DSyncOptions) {
		options.FastUsername = username
		options.FastPassword = password
	}
}

func (WithOption) FastAddresses(addrs ...string) DSyncOption {
	return func(options *DSyncOptions) {
		for _, endpoint := range addrs {
			if _, _, err := net.SplitHostPort(endpoint); err != nil {
				panic(err)
			}
		}
		options.FastAddresses = addrs
	}
}

func (WithOption) FastSecure(secure bool) DSyncOption {
	return func(o *DSyncOptions) {
		o.FastSecure = secure
	}
}

func (WithOption) FastTLSConfig(conf *tls.Config) DSyncOption {
	return func(o *DSyncOptions) {
		o.FastTLSConfig = conf
	}
}
