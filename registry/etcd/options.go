package etcd

import (
	"crypto/tls"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net"
	"strings"
	"time"
)

type EtcdOptions struct {
	EtcdClient    *clientv3.Client
	EtcdConfig    *clientv3.Config
	KeyPrefix     string
	Timeout       time.Duration
	WatchChanSize int
	FastUsername  string
	FastPassword  string
	FastAddresses []string
	FastSecure    bool
	FastTLSConfig *tls.Config
}

type EtcdOption func(options *EtcdOptions)

type WithEtcdOption struct{}

func (WithEtcdOption) Default() EtcdOption {
	return func(options *EtcdOptions) {
		WithEtcdOption{}.EtcdClient(nil)(options)
		WithEtcdOption{}.EtcdConfig(nil)(options)
		WithEtcdOption{}.KeyPrefix("/golaxy/registry/")(options)
		WithEtcdOption{}.Timeout(3 * time.Second)(options)
		WithEtcdOption{}.WatchChanSize(128)(options)
		WithEtcdOption{}.FastAuth("", "")(options)
		WithEtcdOption{}.FastAddresses("127.0.0.1:2379")(options)
		WithEtcdOption{}.FastSecure(false)(options)
		WithEtcdOption{}.FastTLSConfig(nil)(options)
	}
}

func (WithEtcdOption) EtcdClient(cli *clientv3.Client) EtcdOption {
	return func(o *EtcdOptions) {
		o.EtcdClient = cli
	}
}

func (WithEtcdOption) EtcdConfig(config *clientv3.Config) EtcdOption {
	return func(o *EtcdOptions) {
		o.EtcdConfig = config
	}
}

func (WithEtcdOption) KeyPrefix(prefix string) EtcdOption {
	return func(options *EtcdOptions) {
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.KeyPrefix = prefix
	}
}

func (WithEtcdOption) Timeout(dur time.Duration) EtcdOption {
	return func(options *EtcdOptions) {
		options.Timeout = dur
	}
}

func (WithEtcdOption) WatchChanSize(size int) EtcdOption {
	return func(options *EtcdOptions) {
		if size < 0 {
			panic("options.WatchChanSize can't be set to a value less then 0")
		}
		options.WatchChanSize = size
	}
}

func (WithEtcdOption) FastAuth(username, password string) EtcdOption {
	return func(options *EtcdOptions) {
		options.FastUsername = username
		options.FastPassword = password
	}
}

func (WithEtcdOption) FastAddresses(addrs ...string) EtcdOption {
	return func(options *EtcdOptions) {
		for _, endpoint := range addrs {
			if _, _, err := net.SplitHostPort(endpoint); err != nil {
				panic(err)
			}
		}
		options.FastAddresses = addrs
	}
}

func (WithEtcdOption) FastSecure(secure bool) EtcdOption {
	return func(o *EtcdOptions) {
		o.FastSecure = secure
	}
}

func (WithEtcdOption) FastTLSConfig(conf *tls.Config) EtcdOption {
	return func(o *EtcdOptions) {
		o.FastTLSConfig = conf
	}
}
