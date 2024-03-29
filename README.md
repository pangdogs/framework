# framework

Golaxy分布式服务开发框架，基于Actor线程模型（Actor Model）与分布式实体（Distributed Entities）架构，可以快速开发服务器应用。

本框架特点如下：

...

本包提供以下插件：

- 消息队列与事件驱动架构（MQ and Broker），基于NATS。
- 服务发现（Service Discovery），基于Etcd或Redis。
- 分布式锁（Distributed Sync），基于Etcd或Redis。
- GTP协议（Golaxy Transfer Protocol），适用于长连接、实时通信的工作场景，需要工作在可靠网络协议（TCP/KCP）之上，支持链路加密、链路鉴权、断线续连等特性。
- GTP协议网关和客户端（GTP Gate and Client）。
- GAP协议（Golaxy Application Protocol），适用于开发应用层通信消息，需要工作在GTP协议或MQ之上，支持消息判重，解决了幂等性问题。
- 分布式服务支持（Distributed Service）。
- 分布式实体支持（Distributed Entities）。
- RPC支持（Remote Procedure Call）。
- 日志（Logger）。

## Install
```
go get -u git.golaxy.org/framework
```
