package redisdb

import (
	"context"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/plugins/db"
	"git.golaxy.org/framework/plugins/log"
	"github.com/redis/go-redis/v9"
)

type IRedisDB interface {
	RedisDB(tag string) *redis.Client
}

func newRedisDB(settings ...option.Setting[RedisDBOptions]) IRedisDB {
	return &_RedisDB{
		options: option.Make(With.Default(), settings...),
		dbs:     make(map[string]*redis.Client),
	}
}

type _RedisDB struct {
	options RedisDBOptions
	servCtx service.Context
	dbs     map[string]*redis.Client
}

func (r *_RedisDB) InitSP(ctx service.Context) {
	log.Infof(ctx, "init plugin %q", self.Name)

	r.servCtx = ctx

	for _, info := range r.options.DBInfos {
		r.dbs[info.Tag] = r.connectToDB(info)
	}
}

func (r *_RedisDB) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut plugin %q", self.Name)

	for _, db := range r.dbs {
		db.Close()
	}
}

func (r *_RedisDB) RedisDB(tag string) *redis.Client {
	return r.dbs[tag]
}

func (r *_RedisDB) connectToDB(info db.DBInfo) *redis.Client {
	opt, err := redis.ParseURL(info.ConnStr)
	if err != nil {
		log.Panicf(r.servCtx, "parse db conn str %q failed, %v", info.ConnStr, err)
	}

	rdb := redis.NewClient(opt)

	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Panicf(r.servCtx, "ping db %q failed, %s", info.ConnStr, err)
	}

	log.Infof(r.servCtx, "conn to db %q ok", info.ConnStr)
	return rdb
}
