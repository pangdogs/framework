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

package sqldb

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/db/dsn"
	"git.golaxy.org/framework/addins/log"
	"github.com/glebarez/sqlite"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type ISQLDB interface {
	DB(tag string) *gorm.DB
	ReflectedDB(tag string) reflect.Value
}

func DB(svcCtx service.Context, tag string) *gorm.DB {
	return AddIn.Require(svcCtx).DB(tag)
}

func newSQLDB(settings ...option.Setting[SQLDBOptions]) ISQLDB {
	return &_SQLDB{
		options: option.New(With.Default(), settings...),
		dbs:     make(map[string]*_GormDB),
	}
}

type _GormDB struct {
	db        *gorm.DB
	reflected reflect.Value
}

type _SQLDB struct {
	svcCtx  service.Context
	options SQLDBOptions
	dbs     map[string]*_GormDB
}

func (s *_SQLDB) Init(svcCtx service.Context) {
	log.L(svcCtx).Info("initializing add-in", zap.String("name", AddIn.Name))

	s.svcCtx = svcCtx

	for _, info := range s.options.DBInfos {
		db := s.connectToDB(info)

		s.dbs[info.Tag] = &_GormDB{
			db:        db,
			reflected: reflect.ValueOf(db),
		}
	}

	if len(s.dbs) <= 0 {
		log.L(svcCtx).Warn("no sql db has been connected")
	}
}

func (s *_SQLDB) Shut(svcCtx service.Context) {
	log.L(svcCtx).Info("shutting down add-in", zap.String("name", AddIn.Name))

	for _, db := range s.dbs {
		sqldb, _ := db.db.DB()
		if sqldb != nil {
			sqldb.Close()
		}
	}
}

func (s *_SQLDB) DB(tag string) *gorm.DB {
	db := s.dbs[tag]
	if db == nil {
		return nil
	}
	return db.db
}

func (s *_SQLDB) ReflectedDB(tag string) reflect.Value {
	db := s.dbs[tag]
	if db == nil {
		return reflect.Value{}
	}
	return db.reflected
}

func (s *_SQLDB) connectToDB(info *dsn.DBInfo) *gorm.DB {
	dbConnStrUrl, dbConnStrValues, _ := strings.Cut(info.ConnStr, "?")
	queryValues, err := url.ParseQuery(dbConnStrValues)
	if err != nil {
		log.L(s.svcCtx).Panic("parse db conn str failed",
			zap.String("db_type", info.Type),
			zap.String("conn_str", info.ConnStr),
			zap.Error(err))
	}

	maxOpenConns := 10
	maxIdleConns := 5
	connMaxIdleTime := 30 * time.Second
	connMaxLifeTime := 10 * time.Minute

	if queryValues.Has("maxOpenConns") {
		maxOpenConns, _ = strconv.Atoi(queryValues.Get("maxOpenConns"))
		queryValues.Del("maxOpenConns")
	}

	if queryValues.Has("maxIdleConns") {
		maxIdleConns, _ = strconv.Atoi(queryValues.Get("maxIdleConns"))
		queryValues.Del("maxIdleConns")
	}

	if queryValues.Has("connMaxIdleTime") {
		connMaxIdleTime, _ = time.ParseDuration(queryValues.Get("connMaxIdleTime"))
		queryValues.Del("connMaxIdleTime")
	}

	if queryValues.Has("connMaxLifeTime") {
		connMaxLifeTime, _ = time.ParseDuration(queryValues.Get("connMaxLifeTime"))
		queryValues.Del("connMaxLifeTime")
	}

	if !queryValues.Has("parseTime") {
		queryValues.Add("parseTime", "True")
	}

	if !queryValues.Has("loc") {
		queryValues.Add("loc", "Local")
	}

	dbConnStrValues = queryValues.Encode()

	dbConnStr := dbConnStrUrl
	if dbConnStrValues != "" {
		dbConnStr = dbConnStrUrl + "?" + dbConnStrValues
	}

	var dial gorm.Dialector

	switch strings.ToLower(info.Type) {
	case strings.ToLower(dsn.MySQL):
		dial = mysql.Open(dbConnStr)
	case strings.ToLower(dsn.PostgreSQL):
		dial = postgres.Open(dbConnStr)
	case strings.ToLower(dsn.SQLServer):
		dial = sqlserver.Open(dbConnStr)
	case strings.ToLower(dsn.SQLite):
		dial = sqlite.Open(dbConnStr)
	default:
		log.L(s.svcCtx).Panic("conn to db failed",
			zap.String("db_type", info.Type),
			zap.String("conn_str", info.ConnStr),
			zap.Error(err))
	}

	db, err := gorm.Open(dial)
	if err != nil {
		log.L(s.svcCtx).Panic("conn to db failed",
			zap.String("db_type", info.Type),
			zap.String("conn_str", info.ConnStr),
			zap.Error(err))
	}

	sqldb, err := db.DB()
	if err != nil {
		log.L(s.svcCtx).Panic("conn to db failed",
			zap.String("db_type", info.Type),
			zap.String("conn_str", info.ConnStr),
			zap.Error(err))
	}

	sqldb.SetMaxOpenConns(maxOpenConns)
	sqldb.SetMaxIdleConns(maxIdleConns)
	sqldb.SetConnMaxIdleTime(connMaxIdleTime)
	sqldb.SetConnMaxLifetime(connMaxLifeTime)

	if err := sqldb.Ping(); err != nil {
		log.L(s.svcCtx).Panic("ping db failed",
			zap.String("db_type", info.Type),
			zap.String("conn_str", info.ConnStr),
			zap.Error(err))
	}

	log.L(s.svcCtx).Info("connect to db ok",
		zap.String("db_type", info.Type),
		zap.String("conn_str", info.ConnStr))
	return db
}
