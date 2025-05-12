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
	"errors"
	"fmt"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/framework/addins/db/mongodb"
	"git.golaxy.org/framework/addins/db/redisdb"
	"git.golaxy.org/framework/addins/db/sqldb"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
	"reflect"
	"strings"
	"unsafe"
)

var (
	gormDBRT      = reflect.TypeFor[*gorm.DB]()
	redisClientRT = reflect.TypeFor[*redis.Client]()
	mongoClientRT = reflect.TypeFor[*mongo.Client]()
)

func InjectDB(svcCtx service.Context, target any) error {
	return InjectDBRV(svcCtx, reflect.ValueOf(target))
}

func InjectDBRV(svcCtx service.Context, target reflect.Value) error {
	if svcCtx == nil {
		return fmt.Errorf("db: %w: svcCtx is nil", exception.ErrArgs)
	}

	targetRT := target.Type()

retry:
	switch target.Kind() {
	case reflect.Struct:
		for i := range target.NumField() {
			field := targetRT.Field(i)

			switch field.Type {
			case gormDBRT, redisClientRT, mongoClientRT:
				break
			default:
				continue
			}

			tag := strings.TrimSpace(field.Tag.Get("db"))
			if tag == "-" {
				continue
			}

			var db reflect.Value

			switch field.Type {
			case gormDBRT:
				db = sqldb.Using(svcCtx).ReflectedSQLDB(tag)
			case redisClientRT:
				db = redisdb.Using(svcCtx).ReflectedRedisDB(tag)
			case mongoClientRT:
				db = mongodb.Using(svcCtx).ReflectedMongoDB(tag)
			}

			if !db.IsValid() {
				continue
			}

			if field.IsExported() {
				target.Field(i).Set(db)
			} else {
				ptr := unsafe.Pointer(target.Field(i).UnsafeAddr())
				fieldPtr := reflect.NewAt(field.Type, ptr).Elem()
				fieldPtr.Set(db)
			}
		}

		return nil

	case reflect.Pointer, reflect.Interface:
		if target.IsNil() {
			return errors.New("db: target is nil")
		}

		target = target.Elem()
		targetRT = target.Type()

		goto retry

	default:
		return fmt.Errorf("db: invalid taget %s", targetRT.Kind())
	}
}
