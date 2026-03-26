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

package redisdb

import (
	"context"
	"reflect"

	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/db/dsn"
	"git.golaxy.org/framework/addins/log"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type IRedisDB interface {
	DB(tag string) *redis.Client
	ReflectedDB(tag string) reflect.Value
}

func DB(svcCtx service.Context, tag string) *redis.Client {
	return AddIn.Require(svcCtx).DB(tag)
}

func newRedisDB(settings ...option.Setting[RedisDBOptions]) IRedisDB {
	return &_RedisDB{
		options: option.New(With.Default(), settings...),
		dbs:     make(map[string]*_RedisClient),
	}
}

type _RedisClient struct {
	client    *redis.Client
	reflected reflect.Value
}

type _RedisDB struct {
	svcCtx  service.Context
	options RedisDBOptions
	dbs     map[string]*_RedisClient
}

func (r *_RedisDB) Init(svcCtx service.Context) {
	log.L(svcCtx).Info("initializing add-in", zap.String("name", AddIn.Name))

	r.svcCtx = svcCtx

	for _, info := range r.options.DBInfos {
		cli := r.connectToDB(info)

		r.dbs[info.Tag] = &_RedisClient{
			client:    cli,
			reflected: reflect.ValueOf(cli),
		}
	}

	if len(r.dbs) <= 0 {
		log.L(svcCtx).Warn("no sql db has been connected")
	}
}

func (r *_RedisDB) Shut(svcCtx service.Context) {
	log.L(svcCtx).Info("shutting down add-in", zap.String("name", AddIn.Name))

	for _, db := range r.dbs {
		db.client.Close()
	}
}

func (r *_RedisDB) DB(tag string) *redis.Client {
	cli := r.dbs[tag]
	if cli == nil {
		return nil
	}
	return cli.client
}

func (r *_RedisDB) ReflectedDB(tag string) reflect.Value {
	cli := r.dbs[tag]
	if cli == nil {
		return reflect.Value{}
	}
	return cli.reflected
}

func (r *_RedisDB) connectToDB(info *dsn.DBInfo) *redis.Client {
	opt, err := redis.ParseURL(info.ConnStr)
	if err != nil {
		log.L(r.svcCtx).Panic("parse db conn str failed",
			zap.String("db_type", info.Type),
			zap.String("conn_str", info.ConnStr),
			zap.Error(err))
	}

	rdb := redis.NewClient(opt)

	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		log.L(r.svcCtx).Panic("ping db failed",
			zap.String("db_type", info.Type),
			zap.String("conn_str", info.ConnStr),
			zap.Error(err))
	}

	log.L(r.svcCtx).Info("connect to db ok",
		zap.String("db_type", info.Type),
		zap.String("conn_str", info.ConnStr))
	return rdb
}
