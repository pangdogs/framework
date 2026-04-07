// Package sqldb 提供基于 GORM 的 SQL 数据库 add-in。
//
// 可通过 AddIn 安装提供者，通过 With 配置具名 DSN，并通过 DB 从
// service context 中获取指定 tag 的 *gorm.DB。
package sqldb
