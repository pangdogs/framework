// Package log 提供共享的基于 Zap 的日志 add-in。
//
// 可通过 L 和 S 从 add-in provider 中获取 Logger 或 SugaredLogger，并通过
// With 自定义基础日志器。安装到 Service 的实现会保留到服务终止后，因此仍可在
// OnTerminated 回调中使用；安装到 Runtime 时则遵循 Runtime 的正常关闭流程。
package log
