package etcd_dsync

import (
	"crypto/tls"
	etcd_client "go.etcd.io/etcd/client/v3"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util/option"
	"kit.golaxy.org/golaxy/util/types"
	"kit.golaxy.org/plugins/dsync"
	"kit.golaxy.org/plugins/log"
)

func newDSync(settings ...option.Setting[DSyncOptions]) dsync.DSync {
	return &_DSync{
		options: option.Make(Option{}.Default(), settings...),
	}
}

type _DSync struct {
	options DSyncOptions
	servCtx service.Context
	client  *etcd_client.Client
}

// InitSP 初始化服务插件
func (s *_DSync) InitSP(ctx service.Context) {
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
func (s *_DSync) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut service plugin <%s>:[%s]", plugin.Name, types.AnyFullName(*s))

	if s.options.EtcdClient == nil {
		if s.client != nil {
			s.client.Close()
		}
	}
}

// NewMutex returns a new distributed mutex with given name.
func (s *_DSync) NewMutex(name string, settings ...option.Setting[dsync.DMutexOptions]) dsync.DMutex {
	return s.newMutex(name, option.Make(dsync.Option{}.Default(), settings...))
}

// Separator return name path separator.
func (s *_DSync) Separator() string {
	return "/"
}

func (s *_DSync) configure() etcd_client.Config {
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
