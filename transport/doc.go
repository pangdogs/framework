// Package transport 在应用层实现了Golaxy传输协议（golaxy transport protocol），支持链路鉴权、数据加密、断线重连等功能。
// 本协议主要是性能取向，只做了简单的数据传输加密，未解决复杂的中间人攻击问题，对于安全性要求较高的应用场景，应该使用TLS协议直接加密链路，并关闭本协议的数据加密选项。
package transport
