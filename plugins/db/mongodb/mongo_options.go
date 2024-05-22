package mongodb

import (
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/plugins/db"
	"github.com/elliotchance/pie/v2"
)

type MongoDBOptions struct {
	DBInfos []db.DBInfo
}

var With _Option

type _Option struct{}

func (_Option) Default() option.Setting[MongoDBOptions] {
	return func(options *MongoDBOptions) {
		With.DBInfos().Apply(options)
	}
}

func (_Option) DBInfos(infos ...db.DBInfo) option.Setting[MongoDBOptions] {
	return func(options *MongoDBOptions) {
		options.DBInfos = pie.Filter(infos, func(info db.DBInfo) bool {
			switch info.Type {
			case db.MongoDB:
				return true
			}
			return false
		})
	}
}
