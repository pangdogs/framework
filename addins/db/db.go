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

package db

import (
	"git.golaxy.org/core/service"
	"git.golaxy.org/framework/addins/db/mongodb"
	"git.golaxy.org/framework/addins/db/redisdb"
	"git.golaxy.org/framework/addins/db/sqldb"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

func SQLDB(svcCtx service.Context, tag string) *gorm.DB {
	return sqldb.Using(svcCtx).SQLDB(tag)
}

func RedisDB(svcCtx service.Context, tag string) *redis.Client {
	return redisdb.Using(svcCtx).RedisDB(tag)
}

func MongoDB(svcCtx service.Context, tag string) *mongo.Client {
	return mongodb.Using(svcCtx).MongoDB(tag)
}
