package etcd_dsync

import (
	"crypto/tls"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"kit.golaxy.org/golaxy"
	"net"
	"strings"
)

type Option struct{}

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

func (Option) Default() DSyncOption {
	return func(options *DSyncOptions) {
		Option{}.EtcdClient(nil)(options)
		Option{}.EtcdConfig(nil)(options)
		Option{}.KeyPrefix("/golaxy/mutex/")(options)
		Option{}.WatchChanSize(128)(options)
		Option{}.FastAuth("", "")(options)
		Option{}.FastAddresses("127.0.0.1:2379")(options)
		Option{}.FastSecure(false)(options)
		Option{}.FastTLSConfig(nil)(options)
	}
}

func (Option) EtcdClient(cli *clientv3.Client) DSyncOption {
	return func(o *DSyncOptions) {
		o.EtcdClient = cli
	}
}

func (Option) EtcdConfig(config *clientv3.Config) DSyncOption {
	return func(o *DSyncOptions) {
		o.EtcdConfig = config
	}
}

func (Option) KeyPrefix(prefix string) DSyncOption {
	return func(options *DSyncOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.KeyPrefix = prefix
	}
}

func (Option) WatchChanSize(size int) DSyncOption {
	return func(options *DSyncOptions) {
		if size < 0 {
			panic(fmt.Errorf("%w: option WatchChanSize can't be set to a value less then 0", golaxy.ErrArgs))
		}
		options.WatchChanSize = size
	}
}

func (Option) FastAuth(username, password string) DSyncOption {
	return func(options *DSyncOptions) {
		options.FastUsername = username
		options.FastPassword = password
	}
}

func (Option) FastAddresses(addrs ...string) DSyncOption {
	return func(options *DSyncOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				panic(err)
			}
		}
		options.FastAddresses = addrs
	}
}

func (Option) FastSecure(secure bool) DSyncOption {
	return func(o *DSyncOptions) {
		o.FastSecure = secure
	}
}

func (Option) FastTLSConfig(conf *tls.Config) DSyncOption {
	return func(o *DSyncOptions) {
		o.FastTLSConfig = conf
	}
}
