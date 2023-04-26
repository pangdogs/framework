package etcd

import (
	"crypto/tls"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net"
	"strings"
	"time"
)

type EtcdOptions struct {
	EtcdClient *clientv3.Client
	EtcdConfig *clientv3.Config
	KeyPrefix  string
	Timeout    time.Duration
	Username   string
	Password   string
	Addresses  []string
	Secure     bool
	TLSConfig  *tls.Config
}

type EtcdOption func(options *EtcdOptions)

type WithEtcdOption struct{}

func (WithEtcdOption) Default() EtcdOption {
	return func(options *EtcdOptions) {
		WithEtcdOption{}.EtcdClient(nil)(options)
		WithEtcdOption{}.EtcdConfig(nil)(options)
		WithEtcdOption{}.KeyPrefix("/golaxy/registry/")(options)
		WithEtcdOption{}.Timeout(3 * time.Second)(options)
		WithEtcdOption{}.Auth("", "")(options)
		WithEtcdOption{}.Addresses("127.0.0.1:2379")(options)
		WithEtcdOption{}.Secure(false)(options)
		WithEtcdOption{}.TLSConfig(nil)(options)
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
		if dur <= 0 {
			panic("options.Timeout can't be set to a value less equal 0")
		}
		options.Timeout = dur
	}
}

func (WithEtcdOption) Auth(username, password string) EtcdOption {
	return func(options *EtcdOptions) {
		options.Username = username
		options.Password = password
	}
}

func (WithEtcdOption) Addresses(addrs ...string) EtcdOption {
	return func(options *EtcdOptions) {
		for _, endpoint := range addrs {
			if _, _, err := net.SplitHostPort(endpoint); err != nil {
				panic(err)
			}
		}
		options.Addresses = addrs
	}
}

func (WithEtcdOption) Secure(secure bool) EtcdOption {
	return func(o *EtcdOptions) {
		o.Secure = secure
	}
}

func (WithEtcdOption) TLSConfig(conf *tls.Config) EtcdOption {
	return func(o *EtcdOptions) {
		o.TLSConfig = conf
	}
}
