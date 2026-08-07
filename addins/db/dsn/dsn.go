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

package dsn

// MySQL、PostgreSQL、SQLServer 和 SQLite 是 SQL add-in 支持的数据库类型；
// Redis 与 Mongo 分别用于对应的非 SQL add-in。
const (
	MySQL      = "mysql"
	PostgreSQL = "postgresql"
	SQLServer  = "sqlserver"
	SQLite     = "sqlite"
	Redis      = "redis"
	Mongo      = "mongo"
)

// DBInfo 描述一条带 tag 的数据库连接配置。
type DBInfo struct {
	Tag     string `json:"tag,omitempty"`      // Tag 是服务内查询连接时使用的名称，可为空。
	Type    string `json:"type,omitempty"`     // Type 选择数据库后端，应使用本包定义的常量。
	ConnStr string `json:"conn_str,omitempty"` // ConnStr 是交给对应驱动解析的连接字符串。
}
