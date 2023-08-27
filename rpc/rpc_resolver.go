package rpc

// RPCResolver RPC解析器，用于收集加入运行时的实体信息，解析支持RPC调用函数路径（包含实体与组件上的函数），使RPCRouter（RPC路由器）可以正确路由RPC调用。（必须安装在运行时上）
type RPCResolver interface {
}
