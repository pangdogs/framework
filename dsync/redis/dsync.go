package redis

import (
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util"
	"kit.golaxy.org/plugins/dsync"
	"kit.golaxy.org/plugins/logger"
	"log"
)

func newRedisDSync(options ...Option) dsync.DSync {
	opts := Options{}
	WithOption{}.Default()(&opts)

	for i := range options {
		options[i](&opts)
	}

	return &_RedisDsync{
		options: opts,
	}
}

type _RedisDsync struct {
	options Options
	ctx     service.Context
	client  *redis.Client
	*redsync.Redsync
}

// InitSP 初始化服务插件
func (s *_RedisDsync) InitSP(ctx service.Context) {
	logger.Infof(ctx, "init service plugin %q with %q", definePlugin.Name, util.TypeOfAnyFullName(*s))

	s.ctx = ctx

	if s.options.RedisClient == nil {
		s.client = redis.NewClient(s.configure())
	} else {
		s.client = s.options.RedisClient
	}

	_, err := s.client.Ping(ctx).Result()
	if err != nil {
		log.Panicf("ping redis %q failed, %v", s.client, err)
	}

	s.Redsync = redsync.New(goredis.NewPool(s.client))
}

// ShutSP 关闭服务插件
func (s *_RedisDsync) ShutSP(ctx service.Context) {
	logger.Infof(ctx, "shut service plugin %q", definePlugin.Name)

	if s.options.RedisClient == nil {
		if s.client != nil {
			s.client.Close()
		}
	}
}

// NewDMutex returns a new distributed mutex with given name.
func (s *_RedisDsync) NewDMutex(name string, options ...dsync.Option) dsync.DMutex {
	opts := dsync.Options{}
	dsync.WithOption{}.Default()(&opts)

	for i := range options {
		options[i](&opts)
	}

	return newRedisDMutex(s, name, opts)
}

func (s *_RedisDsync) configure() *redis.Options {
	if s.options.RedisConfig != nil {
		return s.options.RedisConfig
	}

	if s.options.RedisURL != "" {
		conf, err := redis.ParseURL(s.options.RedisURL)
		if err != nil {
			logger.Panicf(s.ctx, "parse redis url %q failed, %s", s.options.RedisURL, err)
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
