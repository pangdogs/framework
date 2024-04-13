# framework

Golaxy分布式服务开发框架，基于Actor线程模型（Actor Model）与分布式实体（Distributed Entities）架构，可以快速开发服务器应用。

本框架特点如下：

...

本包提供以下插件：

- 消息队列与事件驱动架构（MQ and Broker），基于NATS。
- 服务发现（Service Discovery），基于Etcd或Redis。
- 分布式锁（Distributed Sync），基于Etcd或Redis。
- 分布式服务支持（Distributed Service），定义分布式节点地址格式，提供异步模型未来（Future），支持分布式服务间通信，可以横向拓展服务。
- 分布式实体支持（Distributed Entities），提供分布式实体信息上报与查询功能，支持分布式实体间通信。
- GTP协议（Golaxy Transfer Protocol），适用于长连接、实时通信的工作场景，需要工作在可靠网络协议（TCP/WebSocket）之上，支持双向签名验证、链路加密、链路鉴权、断线续连重传、自定义消息等特性。
- GTP协议网关与客户端（GTP Gate and Client），基于TCP/WebSocket的GTP网关与客户端实现。
- GAP协议（Golaxy Application Protocol），适用于开发应用层通信消息，需要工作在GTP协议或MQ之上，支持消息判重、自定义消息、自定义可变类型等特性。
- 路由（Router），支持规划路线，映射网关上的会话与实体，使任意服务可以与客户端通信，还可以创建分组，支持消息组播。
- RPC支持（Remote Procedure Call），支持服务间、实体间、或与客户端和分组间进行RPC调用，基于GAP协议，支持可变类型，整体做到了简单易用。
- 日志（Logger），基于Zap。
- 配置（Config），基于Viper。

## Install
```
go get -u git.golaxy.org/framework
```
