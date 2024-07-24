package sqldb

import (
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/plugins/db"
	"git.golaxy.org/framework/plugins/log"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ISQLDB interface {
	SQLDB(tag string) *gorm.DB
}

func newSQLDB(settings ...option.Setting[SQLDBOptions]) ISQLDB {
	return &_SQLDB{
		options: option.Make(With.Default(), settings...),
		dbs:     make(map[string]*gorm.DB),
	}
}

type _SQLDB struct {
	servCtx service.Context
	options SQLDBOptions
	dbs     map[string]*gorm.DB
}

func (s *_SQLDB) InitSP(ctx service.Context) {
	log.Infof(ctx, "init plugin %q", self.Name)

	s.servCtx = ctx

	for _, info := range s.options.DBInfos {
		s.dbs[info.Tag] = s.connectToDB(info)
	}
}

func (s *_SQLDB) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut plugin %q", self.Name)

	for _, db := range s.dbs {
		sqldb, _ := db.DB()
		if sqldb != nil {
			sqldb.Close()
		}
	}
}

func (s *_SQLDB) SQLDB(tag string) *gorm.DB {
	return s.dbs[tag]
}

func (s *_SQLDB) connectToDB(info db.DBInfo) *gorm.DB {
	dbConnStrUrl, dbConnStrValues, _ := strings.Cut(info.ConnStr, "?")
	queryValues, err := url.ParseQuery(dbConnStrValues)
	if err != nil {
		log.Panicf(s.servCtx, "parse db(%s) conn str %q failed, %v", info.Type, info.ConnStr, err)
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
	case strings.ToLower(db.MySQL):
		dial = mysql.Open(dbConnStr)
	case strings.ToLower(db.PostgreSQL):
		dial = postgres.Open(dbConnStr)
	case strings.ToLower(db.SQLServer):
		dial = sqlserver.Open(dbConnStr)
	case strings.ToLower(db.SQLite):
		dial = sqlite.Open(dbConnStr)
	default:
		log.Panicf(s.servCtx, "conn to db(%s) %q failed, not", info.Type, dbConnStr)
	}

	db, err := gorm.Open(dial)
	if err != nil {
		log.Panicf(s.servCtx, "conn to db(%s) %q failed, %s", info.Type, dbConnStr, err)
	}

	sqldb, err := db.DB()
	if err != nil {
		log.Panicf(s.servCtx, "conn to db(%s) %q failed, %s", info.Type, dbConnStr, err)
	}

	sqldb.SetMaxOpenConns(maxOpenConns)
	sqldb.SetMaxIdleConns(maxIdleConns)
	sqldb.SetConnMaxIdleTime(connMaxIdleTime)
	sqldb.SetConnMaxLifetime(connMaxLifeTime)

	if err := sqldb.Ping(); err != nil {
		log.Panicf(s.servCtx, "ping db(%s) %q failed, %s", info.Type, dbConnStr, err)
	}

	log.Infof(s.servCtx, "conn to db(%s) %q ok", info.Type, dbConnStr)
	return db
}
