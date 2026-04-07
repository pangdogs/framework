// Package rpcstack 提供 runtime 作用域的 RPC 调用链上下文。
//
// 该 add-in 保存当前调用链和每次调用关联的变量，供处理器和代理辅助代码
// 在嵌套 RPC 调用间传递请求元数据。
package rpcstack
