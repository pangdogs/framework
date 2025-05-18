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

package mongodb

import (
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/db/dbtypes"
	"github.com/elliotchance/pie/v2"
)

type MongoDBOptions struct {
	DBInfos []*dbtypes.DBInfo
}

var With _Option

type _Option struct{}

func (_Option) Default() option.Setting[MongoDBOptions] {
	return func(options *MongoDBOptions) {
		With.DBInfos().Apply(options)
	}
}

func (_Option) DBInfos(infos ...*dbtypes.DBInfo) option.Setting[MongoDBOptions] {
	return func(options *MongoDBOptions) {
		infos = pie.Filter(infos, func(info *dbtypes.DBInfo) bool {
			if info == nil {
				return false
			}
			switch info.Type {
			case dbtypes.MongoDB:
				return true
			}
			return false
		})

		if len(infos) != len(pie.Map(infos, func(info *dbtypes.DBInfo) string { return info.Tag })) {
			exception.Panicf("db: %w: tags in db infos must be unique", exception.ErrArgs)
		}

		options.DBInfos = infos
	}
}
