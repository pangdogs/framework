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
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/db/dbtypes"
	"git.golaxy.org/framework/addins/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"reflect"
)

type IMongoDB interface {
	MongoDB(tag string) *mongo.Client
	ReflectedMongoDB(tag string) reflect.Value
}

func newMongoDB(settings ...option.Setting[MongoDBOptions]) IMongoDB {
	return &_MongoDB{
		options: option.Make(With.Default(), settings...),
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
	log.Infof(svcCtx, "init addin %q", self.Name)

	m.svcCtx = svcCtx

	for _, info := range m.options.DBInfos {
		cli := m.connectToDB(info)

		m.dbs[info.Tag] = &_MongoClient{
			client:    cli,
			reflected: reflect.ValueOf(cli),
		}
	}
}

func (m *_MongoDB) Shut(svcCtx service.Context) {
	log.Infof(svcCtx, "shut addin %q", self.Name)

	for _, db := range m.dbs {
		db.client.Disconnect(context.Background())
	}
}

func (m *_MongoDB) MongoDB(tag string) *mongo.Client {
	cli := m.dbs[tag]
	if cli == nil {
		return nil
	}
	return cli.client
}

func (m *_MongoDB) ReflectedMongoDB(tag string) reflect.Value {
	cli := m.dbs[tag]
	if cli == nil {
		return reflect.Value{}
	}
	return cli.reflected
}

func (m *_MongoDB) connectToDB(info *dbtypes.DBInfo) *mongo.Client {
	opt := options.Client().ApplyURI(info.ConnStr)

	client, err := mongo.NewClient(opt)
	if err != nil {
		log.Panicf(m.svcCtx, "conn to db %q failed, %s", info.ConnStr, err)
	}

	if err := client.Connect(context.Background()); err != nil {
		log.Panicf(m.svcCtx, "conn to db %q failed, %s", info.ConnStr, err)
	}

	if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
		log.Panicf(m.svcCtx, "ping db %q failed, %s", info.ConnStr, err)
	}

	log.Infof(m.svcCtx, "conn to db %q ok", info.ConnStr)
	return client
}
