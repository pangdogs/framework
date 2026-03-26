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

package mongodb

import (
	"context"
	"reflect"

	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/db/dsn"
	"git.golaxy.org/framework/addins/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

type IMongoDB interface {
	DB(tag string) *mongo.Client
	ReflectedDB(tag string) reflect.Value
}

func DB(svcCtx service.Context, tag string) *mongo.Client {
	return AddIn.Require(svcCtx).DB(tag)
}

func newMongoDB(settings ...option.Setting[MongoDBOptions]) IMongoDB {
	return &_MongoDB{
		options: option.New(With.Default(), settings...),
		dbs:     make(map[string]*_MongoClient),
	}
}

type _MongoClient struct {
	client    *mongo.Client
	reflected reflect.Value
}

type _MongoDB struct {
	svcCtx  service.Context
	options MongoDBOptions
	dbs     map[string]*_MongoClient
}

func (m *_MongoDB) Init(svcCtx service.Context) {
	log.L(svcCtx).Info("initializing add-in", zap.String("name", AddIn.Name))

	m.svcCtx = svcCtx

	for _, info := range m.options.DBInfos {
		cli := m.connectToDB(info)

		m.dbs[info.Tag] = &_MongoClient{
			client:    cli,
			reflected: reflect.ValueOf(cli),
		}
	}

	if len(m.dbs) <= 0 {
		log.L(svcCtx).Warn("no sql db has been connected")
	}
}

func (m *_MongoDB) Shut(svcCtx service.Context) {
	log.L(svcCtx).Info("shutting down add-in", zap.String("name", AddIn.Name))

	for _, db := range m.dbs {
		db.client.Disconnect(context.Background())
	}
}

func (m *_MongoDB) DB(tag string) *mongo.Client {
	cli := m.dbs[tag]
	if cli == nil {
		return nil
	}
	return cli.client
}

func (m *_MongoDB) ReflectedDB(tag string) reflect.Value {
	cli := m.dbs[tag]
	if cli == nil {
		return reflect.Value{}
	}
	return cli.reflected
}

func (m *_MongoDB) connectToDB(info *dsn.DBInfo) *mongo.Client {
	opt := options.Client().ApplyURI(info.ConnStr)

	client, err := mongo.NewClient(opt)
	if err != nil {
		log.L(m.svcCtx).Panic("conn to db failed",
			zap.String("db_type", info.Type),
			zap.String("conn_str", info.ConnStr),
			zap.Error(err))
	}

	if err := client.Connect(context.Background()); err != nil {
		log.L(m.svcCtx).Panic("conn to db failed",
			zap.String("db_type", info.Type),
			zap.String("conn_str", info.ConnStr),
			zap.Error(err))
	}

	if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
		log.L(m.svcCtx).Panic("ping db failed",
			zap.String("db_type", info.Type),
			zap.String("conn_str", info.ConnStr),
			zap.Error(err))
	}

	log.L(m.svcCtx).Info("connect to db ok",
		zap.String("db_type", info.Type),
		zap.String("conn_str", info.ConnStr))
	return client
}
