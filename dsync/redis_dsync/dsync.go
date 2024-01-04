package redis_dsync

import (
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util/option"
	"kit.golaxy.org/golaxy/util/types"
	"kit.golaxy.org/plugins/dsync"
	"kit.golaxy.org/plugins/log"
)

func newDSync(settings ...option.Setting[DSyncOptions]) dsync.DSync {
	return &_Dsync{
		options: option.Make(Option{}.Default(), settings...),
	}
}

type _Dsync struct {
	options DSyncOptions
	servCtx service.Context
	client  *redis.Client
	redSync *redsync.Redsync
}

// InitSP 初始化服务插件
func (s *_Dsync) InitSP(ctx service.Context) {
	log.Infof(ctx, "init service plugin <%s>:[%s]", plugin.Name, types.AnyFullName(*s))

	s.servCtx = ctx

	if s.options.RedisClient == nil {
		s.client = redis.NewClient(s.configure())
	} else {
		s.client = s.options.RedisClient
	}

	_, err := s.client.Ping(ctx).Result()
	if err != nil {
		log.Panicf(ctx, "ping redis %q failed, %v", s.client, err)
	}

	s.redSync = redsync.New(goredis.NewPool(s.client))
}

// ShutSP 关闭服务插件
func (s *_Dsync) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut service plugin <%s>:[%s]", plugin.Name, types.AnyFullName(*s))

	if s.options.RedisClient == nil {
		if s.client != nil {
			s.client.Close()
		}
	}
}

// NewMutex returns a new distributed mutex with given name.
func (s *_Dsync) NewMutex(name string, settings ...option.Setting[dsync.DMutexOptions]) dsync.DMutex {
	return s.newMutex(name, option.Make(dsync.Option{}.Default(), settings...))
}

// Separator return name path separator.
func (s *_Dsync) Separator() string {
	return ":"
}

func (s *_Dsync) configure() *redis.Options {
	if s.options.RedisConfig != nil {
		return s.options.RedisConfig
	}

	if s.options.RedisURL != "" {
		conf, err := redis.ParseURL(s.options.RedisURL)
		if err != nil {
			log.Panicf(s.servCtx, "parse redis url %q failed, %s", s.options.RedisURL, err)
		}
		return conf
	}

	conf := &redis.Options{}
	conf.Username = s.options.FastUsername
	conf.Password = s.options.FastPassword
	conf.Addr = s.options.FastAddress
	conf.DB = s.options.FastDB

	return conf
}
