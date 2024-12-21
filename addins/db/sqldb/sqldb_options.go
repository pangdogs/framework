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
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/db"
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
