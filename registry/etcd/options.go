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
	ZapConfig  *zap.Config
	EtcdConfig *clientv3.Config
}

type WithEtcdOption func(options *EtcdOptions)

var EtcdOption = _EtcdOption{}

type _EtcdOption struct{}

func (_EtcdOption) Default() WithEtcdOption {
	return func(options *EtcdOptions) {
		EtcdOption.Auth("", "")(options)
		EtcdOption.Endpoints("127.0.0.1:2379")(options)
		EtcdOption.Timeout(5 * time.Second)(options)
		EtcdOption.Secure(false)(options)
		EtcdOption.TLSConfig(nil)(options)
		EtcdOption.ZapLogger(nil)(options)
		EtcdOption.ZapConfig(nil)(options)
		EtcdOption.EtcdConfig(nil)(options)
	}
}

func (_EtcdOption) Auth(username, password string) WithEtcdOption {
	return func(options *EtcdOptions) {
		options.Username = username
		options.Password = password
	}
}

func (_EtcdOption) Endpoints(endpoints ...string) WithEtcdOption {
	return func(options *EtcdOptions) {
		for _, endpoint := range endpoints {
			if _, _, err := net.SplitHostPort(endpoint); err != nil {
				panic(err)
			}
		}
		options.Endpoints = endpoints
	}
}

func (_EtcdOption) Timeout(dur time.Duration) WithEtcdOption {
	return func(options *EtcdOptions) {
		options.Timeout = dur
	}
}

func (_EtcdOption) Secure(secure bool) WithEtcdOption {
	return func(o *EtcdOptions) {
		o.Secure = secure
	}
}

func (_EtcdOption) TLSConfig(config *tls.Config) WithEtcdOption {
	return func(o *EtcdOptions) {
		o.TLSConfig = config
	}
}

func (_EtcdOption) ZapLogger(zapLogger *zap.Logger) WithEtcdOption {
	return func(o *EtcdOptions) {
		o.ZapLogger = zapLogger
	}
}

func (_EtcdOption) ZapConfig(zapConfig *zap.Config) WithEtcdOption {
	return func(o *EtcdOptions) {
		o.ZapConfig = zapConfig
	}
}

func (_EtcdOption) EtcdConfig(config *clientv3.Config) WithEtcdOption {
	return func(o *EtcdOptions) {
		o.EtcdConfig = config
	}
}
