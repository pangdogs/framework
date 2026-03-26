/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package dsync_redis

import (
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/dsync"
	"git.golaxy.org/framework/addins/log"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func newRedisSync(settings ...option.Setting[RedisSyncOptions]) dsync.IDistSync {
	return &_RedisSync{
		options: option.New(With.Default(), settings...),
	}
}

type _RedisSync struct {
	svcCtx  service.Context
	options RedisSyncOptions
	client  *redis.Client
	redSync *redsync.Redsync
}

// Init 初始化插件
func (s *_RedisSync) Init(svcCtx service.Context) {
	log.L(svcCtx).Info("initializing add-in", zap.String("name", AddIn.Name))

	s.svcCtx = svcCtx

	if s.options.RedisClient == nil {
		s.client = redis.NewClient(s.configure())
	} else {
		s.client = s.options.RedisClient
	}

	_, err := s.client.Ping(svcCtx).Result()
	if err != nil {
		log.L(svcCtx).Panic("ping redis failed", zap.String("db_info", s.client.String()), zap.Error(err))
	}

	s.redSync = redsync.New(goredis.NewPool(s.client))
}

// Shut 关闭插件
func (s *_RedisSync) Shut(svcCtx service.Context) {
	log.L(svcCtx).Info("shutting down add-in", zap.String("name", AddIn.Name))

	if s.options.RedisClient == nil {
		if s.client != nil {
			s.client.Close()
		}
	}
}

// NewMutex 创建分布式锁
func (s *_RedisSync) NewMutex(name string, settings ...option.Setting[dsync.DistMutexOptions]) dsync.IDistMutex {
	return s.newMutex(name, option.New(dsync.With.Default(), settings...))
}

// Separator 获取分隔符
func (s *_RedisSync) Separator() string {
	return ":"
}

func (s *_RedisSync) configure() *redis.Options {
	if s.options.RedisConfig != nil {
		return s.options.RedisConfig
	}

	if s.options.RedisURL != "" {
		conf, err := redis.ParseURL(s.options.RedisURL)
		if err != nil {
			log.L(s.svcCtx).Panic("parse redis url failed", zap.String("url", s.options.RedisURL), zap.Error(err))
		}
		return conf
	}

	conf := &redis.Options{}
	conf.Username = s.options.CustomUsername
	conf.Password = s.options.CustomPassword
	conf.Addr = s.options.CustomAddress
	conf.DB = s.options.CustomDB

	return conf
}
