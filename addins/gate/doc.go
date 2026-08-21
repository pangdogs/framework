// Package gate 提供构建在 GTP 传输栈之上的网关 add-in。
//
// 它为接收外部客户端连接的服务管理监听器、握手、会话以及会话级数据与
// 事件 I/O。会话同时提供绑定自身生命周期的 Context 与异步 Scope。
package gate
