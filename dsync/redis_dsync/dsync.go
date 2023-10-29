package redis_dsync

import (
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util/types"
	"kit.golaxy.org/plugins/dsync"
	"kit.golaxy.org/plugins/log"
)

func newDSync(options ...DSyncOption) dsync.DSync {
	opts := DSyncOptions{}
	Option{}.Default()(&opts)

	for i := range options {
		options[i](&opts)
	}

	return &_Dsync{
		options: opts,
	}
}

type _Dsync struct {
	options DSyncOptions
	ctx     service.Context
	client  *redis.Client
	redSync *redsync.Redsync
}

// InitSP 初始化服务插件
func (s *_Dsync) InitSP(ctx service.Context) {
	log.Infof(ctx, "init service plugin %q with %q", plugin.Name, types.AnyFullName(*s))

	s.ctx = ctx

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
	log.Infof(ctx, "shut service plugin %q", plugin.Name)

	if s.options.RedisClient == nil {
		if s.client != nil {
			s.client.Close()
		}
	}
}

// NewMutex returns a new distributed mutex with given name.
func (s *_Dsync) NewMutex(name string, options ...dsync.DMutexOption) dsync.DMutex {
	opts := dsync.DMutexOptions{}
	dsync.Option{}.Default()(&opts)

	for i := range options {
		options[i](&opts)
	}

	return s.newMutex(name, opts)
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
			log.Panicf(s.ctx, "parse redis url %q failed, %s", s.options.RedisURL, err)
		}
		return conf
	}

	conf := &redis.Options{}
	conf.Username = s.options.FastUsername
	conf.Password = s.options.FastPassword
	conf.Addr = s.options.FastAddress
	conf.DB = s.options.FastDBIndex

	return conf
}
