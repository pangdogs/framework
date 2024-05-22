package sqldb

import (
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/plugins/db"
	"github.com/elliotchance/pie/v2"
)

type SQLDBOptions struct {
	DBInfos []db.DBInfo
}

var With _Option

type _Option struct{}

func (_Option) Default() option.Setting[SQLDBOptions] {
	return func(options *SQLDBOptions) {
		With.DBInfos().Apply(options)
	}
}

func (_Option) DBInfos(infos ...db.DBInfo) option.Setting[SQLDBOptions] {
	return func(options *SQLDBOptions) {
		options.DBInfos = pie.Filter(infos, func(info db.DBInfo) bool {
			switch info.Type {
			case db.MySQL, db.PostgreSQL, db.SQLServer, db.SQLite:
				return true
			}
			return false
		})
	}
}
