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
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/db/dbtypes"
	"git.golaxy.org/framework/addins/log"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type ISQLDB interface {
	SQLDB(tag string) *gorm.DB
	ReflectedSQLDB(tag string) reflect.Value
}

func newSQLDB(settings ...option.Setting[SQLDBOptions]) ISQLDB {
	return &_SQLDB{
		options: option.Make(With.Default(), settings...),
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
	log.Infof(svcCtx, "init addin %q", self.Name)

	s.svcCtx = svcCtx

	for _, info := range s.options.DBInfos {
		db := s.connectToDB(info)

		s.dbs[info.Tag] = &_GormDB{
			db:        db,
			reflected: reflect.ValueOf(db),
		}
	}

	if len(s.dbs) <= 0 {
		log.Warn(svcCtx, "no sql db has been connected")
	}
}

func (s *_SQLDB) Shut(svcCtx service.Context) {
	log.Infof(svcCtx, "shut addin %q", self.Name)

	for _, db := range s.dbs {
		sqldb, _ := db.db.DB()
		if sqldb != nil {
			sqldb.Close()
		}
	}
}

func (s *_SQLDB) SQLDB(tag string) *gorm.DB {
	db := s.dbs[tag]
	if db == nil {
		return nil
	}
	return db.db
}

func (s *_SQLDB) ReflectedSQLDB(tag string) reflect.Value {
	db := s.dbs[tag]
	if db == nil {
		return reflect.Value{}
	}
	return db.reflected
}

func (s *_SQLDB) connectToDB(info *dbtypes.DBInfo) *gorm.DB {
	dbConnStrUrl, dbConnStrValues, _ := strings.Cut(info.ConnStr, "?")
	queryValues, err := url.ParseQuery(dbConnStrValues)
	if err != nil {
		log.Panicf(s.svcCtx, "parse db(%s) conn str %q failed, %v", info.Type, info.ConnStr, err)
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
	case strings.ToLower(dbtypes.MySQL):
		dial = mysql.Open(dbConnStr)
	case strings.ToLower(dbtypes.PostgreSQL):
		dial = postgres.Open(dbConnStr)
	case strings.ToLower(dbtypes.SQLServer):
		dial = sqlserver.Open(dbConnStr)
	case strings.ToLower(dbtypes.SQLite):
		dial = sqlite.Open(dbConnStr)
	default:
		log.Panicf(s.svcCtx, "conn to db(%s) %q failed, not", info.Type, dbConnStr)
	}

	db, err := gorm.Open(dial)
	if err != nil {
		log.Panicf(s.svcCtx, "conn to db(%s) %q failed, %s", info.Type, dbConnStr, err)
	}

	sqldb, err := db.DB()
	if err != nil {
		log.Panicf(s.svcCtx, "conn to db(%s) %q failed, %s", info.Type, dbConnStr, err)
	}

	sqldb.SetMaxOpenConns(maxOpenConns)
	sqldb.SetMaxIdleConns(maxIdleConns)
	sqldb.SetConnMaxIdleTime(connMaxIdleTime)
	sqldb.SetConnMaxLifetime(connMaxLifeTime)

	if err := sqldb.Ping(); err != nil {
		log.Panicf(s.svcCtx, "ping db(%s) %q failed, %s", info.Type, dbConnStr, err)
	}

	log.Infof(s.svcCtx, "conn to db(%s) %q ok", info.Type, dbConnStr)
	return db
}
