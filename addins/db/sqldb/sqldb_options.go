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
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/db/dsn"
	"github.com/elliotchance/pie/v2"
)

// SQLDBOptions 配置需要在 add-in 启动时连接的 SQL 数据库。
type SQLDBOptions struct {
	DBInfos []*dsn.DBInfo // DBInfos 仅保留受支持 SQL 类型的非 nil 项。
}

// With 提供 SQL add-in 的 Option 构造方法。
var With _SQLDBOption

type _SQLDBOption struct{}

// Default 返回不连接任何 SQL 数据库的默认设置。
func (_SQLDBOption) Default() option.Setting[SQLDBOptions] {
	return func(options *SQLDBOptions) {
		With.DBInfos().Apply(options)
	}
}

// DBInfos 设置具名连接列表；nil 项和不受支持的数据库类型会被过滤。
func (_SQLDBOption) DBInfos(infos ...*dsn.DBInfo) option.Setting[SQLDBOptions] {
	return func(options *SQLDBOptions) {
		infos = pie.Filter(infos, func(info *dsn.DBInfo) bool {
			if info == nil {
				return false
			}
			switch info.Type {
			case dsn.MySQL, dsn.PostgreSQL, dsn.SQLServer, dsn.SQLite:
				return true
			}
			return false
		})

		if len(infos) != len(pie.Map(infos, func(info *dsn.DBInfo) string { return info.Tag })) {
			exception.Panicf("db: %w: tags in db infos must be unique", exception.ErrArgs)
		}

		options.DBInfos = infos
	}
}
