// Package dsync 定义分布式同步能力的抽象层，供不同锁实现复用。
//
// 可通过 AddIn 从 service context 中获取 IDistSync，并通过 With 配置
// 锁的过期时间、重试策略等通用行为。
package dsync
