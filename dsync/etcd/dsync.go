package etcd

import (
	"crypto/tls"
	etcd_client "go.etcd.io/etcd/client/v3"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util"
	"kit.golaxy.org/plugins/dsync"
	"kit.golaxy.org/plugins/logger"
)

func newEtcdDSync(options ...Option) dsync.DSync {
	opts := Options{}
	WithOption{}.Default()(&opts)

	for i := range options {
		options[i](&opts)
	}

	return &_EtcdDSync{
		options: opts,
	}
}

type _EtcdDSync struct {
	options Options
	ctx     service.Context
	client  *etcd_client.Client
}

// InitSP 初始化服务插件
func (s *_EtcdDSync) InitSP(ctx service.Context) {
	logger.Infof(ctx, "init service plugin %q with %q", definePlugin.Name, util.TypeOfAnyFullName(*s))

	s.ctx = ctx

	if s.options.EtcdClient == nil {
		cli, err := etcd_client.New(s.configure())
		if err != nil {
			logger.Panic(ctx, err)
		}
		s.client = cli
	} else {
		s.client = s.options.EtcdClient
	}

	for _, ep := range s.client.Endpoints() {
		if _, err := s.client.Status(ctx, ep); err != nil {
			logger.Panicf(ctx, "status etcd %q failed, %s", ep, err)
		}
	}
}

// ShutSP 关闭服务插件
func (s *_EtcdDSync) ShutSP(ctx service.Context) {
	logger.Infof(ctx, "shut service plugin %q", definePlugin.Name)

	if s.options.EtcdClient == nil {
		if s.client != nil {
			s.client.Close()
		}
	}
}

// NewDMutex returns a new distributed mutex with given name.
func (s *_EtcdDSync) NewDMutex(name string, options ...dsync.Option) dsync.DMutex {
	opts := dsync.Options{}
	dsync.WithOption{}.Default()(&opts)

	for i := range options {
		options[i](&opts)
	}

	return newEtcdDMutex(s, name, opts)
}

func (s *_EtcdDSync) configure() etcd_client.Config {
	if s.options.EtcdConfig != nil {
		return *s.options.EtcdConfig
	}

	config := etcd_client.Config{
		Endpoints: s.options.FastAddresses,
		Username:  s.options.FastUsername,
		Password:  s.options.FastPassword,
	}

	if s.options.FastSecure || s.options.FastTLSConfig != nil {
		tlsConfig := s.options.FastTLSConfig
		if tlsConfig == nil {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		config.TLS = tlsConfig
	}

	return config
}
