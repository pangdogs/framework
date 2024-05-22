package mongodb

import (
	"context"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/plugins/db"
	"git.golaxy.org/framework/plugins/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type IMongoDB interface {
	MongoDB(tag string) *mongo.Client
}

func newMongoDB(settings ...option.Setting[MongoDBOptions]) IMongoDB {
	return &_MongoDB{
		options: option.Make(With.Default(), settings...),
		dbs:     make(map[string]*mongo.Client),
	}
}

type _MongoDB struct {
	options MongoDBOptions
	servCtx service.Context
	dbs     map[string]*mongo.Client
}

func (m *_MongoDB) InitSP(ctx service.Context) {
	log.Infof(ctx, "init plugin %q", self.Name)

	m.servCtx = ctx

	for _, info := range m.options.DBInfos {
		m.dbs[info.Tag] = m.connectToDB(info)
	}
}

func (m *_MongoDB) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut plugin %q", self.Name)

	for _, db := range m.dbs {
		db.Disconnect(context.Background())
	}
}

func (m *_MongoDB) MongoDB(tag string) *mongo.Client {
	return m.dbs[tag]
}

func (m *_MongoDB) connectToDB(info db.DBInfo) *mongo.Client {
	opt := options.Client().ApplyURI(info.ConnStr)

	client, err := mongo.NewClient(opt)
	if err != nil {
		log.Panicf(m.servCtx, "conn to db %q failed, %s", info.ConnStr, err)
	}

	if err := client.Connect(context.Background()); err != nil {
		log.Panicf(m.servCtx, "conn to db %q failed, %s", info.ConnStr, err)
	}

	if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
		log.Panicf(m.servCtx, "ping db %q failed, %s", info.ConnStr, err)
	}

	log.Infof(m.servCtx, "conn to db %q ok", info.ConnStr)
	return client
}
