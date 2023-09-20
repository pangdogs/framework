package etcd_registry

import (
	"crypto/tls"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"kit.golaxy.org/golaxy"
	"net"
	"strings"
)

type Option struct{}

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

func (Option) Default() RegistryOption {
	return func(options *RegistryOptions) {
		Option{}.EtcdClient(nil)(options)
		Option{}.EtcdConfig(nil)(options)
		Option{}.KeyPrefix("/golaxy/registry/")(options)
		Option{}.WatchChanSize(128)(options)
		Option{}.FastAuth("", "")(options)
		Option{}.FastAddresses("127.0.0.1:2379")(options)
		Option{}.FastSecure(false)(options)
		Option{}.FastTLSConfig(nil)(options)
	}
}

func (Option) EtcdClient(cli *clientv3.Client) RegistryOption {
	return func(o *RegistryOptions) {
		o.EtcdClient = cli
	}
}

func (Option) EtcdConfig(config *clientv3.Config) RegistryOption {
	return func(o *RegistryOptions) {
		o.EtcdConfig = config
	}
}

func (Option) KeyPrefix(prefix string) RegistryOption {
	return func(options *RegistryOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.KeyPrefix = prefix
	}
}

func (Option) WatchChanSize(size int) RegistryOption {
	return func(options *RegistryOptions) {
		if size < 0 {
			panic(fmt.Errorf("%w: option WatchChanSize can't be set to a value less then 0", golaxy.ErrArgs))
		}
		options.WatchChanSize = size
	}
}

func (Option) FastAuth(username, password string) RegistryOption {
	return func(options *RegistryOptions) {
		options.FastUsername = username
		options.FastPassword = password
	}
}

func (Option) FastAddresses(addrs ...string) RegistryOption {
	return func(options *RegistryOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				panic(fmt.Errorf("%w: %w", golaxy.ErrArgs, err))
			}
		}
		options.FastAddresses = addrs
	}
}

func (Option) FastSecure(secure bool) RegistryOption {
	return func(o *RegistryOptions) {
		o.FastSecure = secure
	}
}

func (Option) FastTLSConfig(conf *tls.Config) RegistryOption {
	return func(o *RegistryOptions) {
		o.FastTLSConfig = conf
	}
}
