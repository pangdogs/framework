// Package db 为服务代码提供跨数据库的辅助能力。
//
// 它通过 InjectDB 串联 sqldb、redisdb 和 mongodb add-in，并通过
// MigrateDB 调用服务暴露的迁移钩子。
package db
