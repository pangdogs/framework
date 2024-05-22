package dbutil

import (
	"git.golaxy.org/core/service"
	"git.golaxy.org/framework/plugins/db/mongodb"
	"git.golaxy.org/framework/plugins/db/redisdb"
	"git.golaxy.org/framework/plugins/db/sqldb"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

func SQLDB(ctx service.Context, tag string) *gorm.DB {
	return sqldb.Using(ctx).SQLDB(tag)
}

func RedisDB(ctx service.Context, tag string) *redis.Client {
	return redisdb.Using(ctx).RedisDB(tag)
}

func MongoDB(ctx service.Context, tag string) *mongo.Client {
	return mongodb.Using(ctx).MongoDB(tag)
}
