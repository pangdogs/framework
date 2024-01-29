package etcd_dsync

import (
	"crypto/tls"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/plugins/dsync"
	"git.golaxy.org/framework/plugins/log"
	etcd_client "go.etcd.io/etcd/client/v3"
	"time"
)

func newDSync(settings ...option.Setting[DSyncOptions]) dsync.IDistSync {
	return &_DistSync{
		options: option.Make(Option{}.Default(), settings...),
	}
}

type _DistSync struct {
	options DSyncOptions
	servCtx service.Context
	client  *etcd_client.Client
}

// InitSP 初始化服务插件
func (s *_DistSync) InitSP(ctx service.Context) {
	log.Infof(ctx, "init plugin %q", plugin.Name)

	s.servCtx = ctx

	if s.options.EtcdClient == nil {
		cli, err := etcd_client.New(s.configure())
		if err != nil {
			log.Panicf(ctx, "new etcd client failed, %s", err)
		}
		s.client = cli
	} else {
		s.client = s.options.EtcdClient
	}

	for _, ep := range s.client.Endpoints() {
		if _, err := s.client.Status(ctx, ep); err != nil {
			log.Panicf(ctx, "status etcd %q failed, %s", ep, err)
		}
	}
}

// ShutSP 关闭服务插件
func (s *_DistSync) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut plugin %q", plugin.Name)

	if s.options.EtcdClient == nil {
		if s.client != nil {
			s.client.Close()
		}
	}
}

// NewMutex returns a new distributed mutex with given name.
func (s *_DistSync) NewMutex(name string, settings ...option.Setting[dsync.DistMutexOptions]) dsync.IDistMutex {
	return s.newMutex(name, option.Make(dsync.Option{}.Default(), settings...))
}

// GetSeparator return name path separator.
func (s *_DistSync) GetSeparator() string {
	return "/"
}

func (s *_DistSync) configure() etcd_client.Config {
	if s.options.EtcdConfig != nil {
		return *s.options.EtcdConfig
	}

	config := etcd_client.Config{
		Endpoints:   s.options.CustomAddresses,
		Username:    s.options.CustomUsername,
		Password:    s.options.CustomPassword,
		DialTimeout: 3 * time.Second,
	}

	if s.options.CustomSecure || s.options.CustomTLSConfig != nil {
		tlsConfig := s.options.CustomTLSConfig
		if tlsConfig == nil {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		config.TLS = tlsConfig
	}

	return config
}
