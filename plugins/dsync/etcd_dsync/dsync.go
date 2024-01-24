package etcd_dsync

import (
	"crypto/tls"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/core/util/types"
	"git.golaxy.org/framework/plugins/dsync"
	"git.golaxy.org/framework/plugins/log"
	etcd_client "go.etcd.io/etcd/client/v3"
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
	log.Infof(ctx, "init service plugin <%s>:[%s]", plugin.Name, types.AnyFullName(*s))

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
	log.Infof(ctx, "shut service plugin <%s>:[%s]", plugin.Name, types.AnyFullName(*s))

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
		Endpoints: s.options.CustAddresses,
		Username:  s.options.CustUsername,
		Password:  s.options.CustPassword,
	}

	if s.options.CustSecure || s.options.CustTLSConfig != nil {
		tlsConfig := s.options.CustTLSConfig
		if tlsConfig == nil {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		config.TLS = tlsConfig
	}

	return config
}
