package etcd

import (
	"crypto/tls"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"net"
	"time"
)

type EtcdOptions struct {
	Username   string
	Password   string
	Endpoints  []string
	Timeout    time.Duration
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
		WithEtcdOption{}.Endpoints("127.0.0.1:2379")(options)
		WithEtcdOption{}.Timeout(5 * time.Second)(options)
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

func (WithEtcdOption) Endpoints(endpoints ...string) EtcdOption {
	return func(options *EtcdOptions) {
		for _, endpoint := range endpoints {
			if _, _, err := net.SplitHostPort(endpoint); err != nil {
				panic(err)
			}
		}
		options.Endpoints = endpoints
	}
}

func (WithEtcdOption) Timeout(dur time.Duration) EtcdOption {
	return func(options *EtcdOptions) {
		options.Timeout = dur
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
