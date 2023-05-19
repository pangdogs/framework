package etcd

import (
	"crypto/tls"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net"
	"strings"
)

type Options struct {
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

type Option func(options *Options)

type WithOption struct{}

func (WithOption) Default() Option {
	return func(options *Options) {
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

func (WithOption) EtcdClient(cli *clientv3.Client) Option {
	return func(o *Options) {
		o.EtcdClient = cli
	}
}

func (WithOption) EtcdConfig(config *clientv3.Config) Option {
	return func(o *Options) {
		o.EtcdConfig = config
	}
}

func (WithOption) KeyPrefix(prefix string) Option {
	return func(options *Options) {
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.KeyPrefix = prefix
	}
}

func (WithOption) WatchChanSize(size int) Option {
	return func(options *Options) {
		if size < 0 {
			panic("option WatchChanSize can't be set to a value less then 0")
		}
		options.WatchChanSize = size
	}
}

func (WithOption) FastAuth(username, password string) Option {
	return func(options *Options) {
		options.FastUsername = username
		options.FastPassword = password
	}
}

func (WithOption) FastAddresses(addrs ...string) Option {
	return func(options *Options) {
		for _, endpoint := range addrs {
			if _, _, err := net.SplitHostPort(endpoint); err != nil {
				panic(err)
			}
		}
		options.FastAddresses = addrs
	}
}

func (WithOption) FastSecure(secure bool) Option {
	return func(o *Options) {
		o.FastSecure = secure
	}
}

func (WithOption) FastTLSConfig(conf *tls.Config) Option {
	return func(o *Options) {
		o.FastTLSConfig = conf
	}
}
