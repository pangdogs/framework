package etcd

import (
	"crypto/tls"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"net"
	"strings"
	"time"
)

type EtcdOptions struct {
	Username   string
	Password   string
	Addresses  []string
	Timeout    time.Duration
	KeyPrefix  string
	Secure     bool
	TLSConfig  *tls.Config
	ZapLogger  *zap.Logger
	EtcdConfig *clientv3.Config
}

type EtcdOption func(options *EtcdOptions)

type WithEtcdOption struct{}

func (WithEtcdOption) Default() EtcdOption {
	return func(options *EtcdOptions) {
		WithEtcdOption{}.Auth("", "")(options)
		WithEtcdOption{}.Addresses("127.0.0.1:2379")(options)
		WithEtcdOption{}.Timeout(5 * time.Second)(options)
		WithEtcdOption{}.KeyPrefix("/golaxy/registry/")(options)
		WithEtcdOption{}.Secure(false)(options)
		WithEtcdOption{}.TLSConfig(nil)(options)
		WithEtcdOption{}.ZapLogger(nil)(options)
		WithEtcdOption{}.EtcdConfig(nil)(options)
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

func (WithEtcdOption) Timeout(dur time.Duration) EtcdOption {
	return func(options *EtcdOptions) {
		options.Timeout = dur
	}
}

func (WithEtcdOption) KeyPrefix(v string) EtcdOption {
	return func(options *EtcdOptions) {
		if !strings.HasSuffix(v, "/") {
			v += "/"
		}
		options.KeyPrefix = v
	}
}

func (WithEtcdOption) Secure(secure bool) EtcdOption {
	return func(o *EtcdOptions) {
		o.Secure = secure
	}
}

func (WithEtcdOption) TLSConfig(config *tls.Config) EtcdOption {
	return func(o *EtcdOptions) {
		o.TLSConfig = config
	}
}

func (WithEtcdOption) ZapLogger(zapLogger *zap.Logger) EtcdOption {
	return func(o *EtcdOptions) {
		o.ZapLogger = zapLogger
	}
}

func (WithEtcdOption) EtcdConfig(config *clientv3.Config) EtcdOption {
	return func(o *EtcdOptions) {
		o.EtcdConfig = config
	}
}
