package redisdb

import (
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/plugins/db"
	"github.com/elliotchance/pie/v2"
)

type RedisDBOptions struct {
	DBInfos []db.DBInfo
}

var With _Option

type _Option struct{}

func (_Option) Default() option.Setting[RedisDBOptions] {
	return func(options *RedisDBOptions) {
		With.DBInfos().Apply(options)
	}
}

func (_Option) DBInfos(infos ...db.DBInfo) option.Setting[RedisDBOptions] {
	return func(options *RedisDBOptions) {
		options.DBInfos = pie.Filter(infos, func(info db.DBInfo) bool {
			switch info.Type {
			case db.Redis:
				return true
			}
			return false
		})
	}
}
