// Package rpc 提供服务侧 RPC add-in，以及配套的代理与结果辅助能力。
//
// 可通过 AddIn 安装传输实现，通过 ProxyService、ProxyRuntime、ProxyEntity
// 构建调用入口，并通过 Result 或 Assert 辅助函数解析返回的 future。
package rpc
