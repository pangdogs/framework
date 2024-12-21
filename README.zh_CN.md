# Framework
[English](./README.md) | [简体中文](./README.zh_CN.md)

## 简介
[**Golaxy分布式服务开发框架**](https://github.com/pangdogs/framework) 旨在为实时通信应用程序提供一个全面的服务端解决方案。基于EC系统与Actor线程模型的 [**内核**](https://github.com/pangdogs/core) ，框架实现了分布式服务的所有依赖项，设计简洁、易于使用，特别适合用于开发游戏和远程控制系统。

## 功能与特性
基于框架可以开发有状态（`Stateful`）或无状态（`Stateless`）分布式服务，已实现功能与特性：

- 消息队列与事件驱动架构（`MQ and Broker`）：基于NATS。
- 服务发现（`Service Discovery`）：基于ETCD，框架也提供Redis支持，但是Redis没有数据版本控制机制，所以仅能用于功能演示。
- 分布式锁（`Distributed Synchronization`）：基于ETCD或Redis，默认ETCD。
- 分布式服务支持（`Distributed Service`）：定义分布式服务节点地址格式，提供异步模型未来（`Future`），支持分布式服务间通信，可以横向拓展服务。
- 分布式实体支持（`Distributed Entities`）：提供分布式实体（`Distributed Entity`）信息登记与查询功能，支持分布式实体间通信。
- GTP协议（`Golaxy Transfer Protocol`）：适用于长连接、实时通信的工作场景，需要工作在可靠网络协议（`TCP/WebSocket`）之上，支持双向签名验证、链路加密、链路鉴权、断线续连重传、自定义消息等特性。
- GAP协议（`Golaxy Application Protocol`）：适用于开发应用层通信消息，需要工作在`GTP协议`或`MQ`之上，支持消息判重、自定义消息、自定义可变类型等特性。
- GTP协议网关与客户端（`GTP Gate and Client`）：基于`GTP协议`的网关与客户端，支持`TCP/WebSocket`长连接。
- 路由（`Router`）：支持规划通信路线，映射会话与实体，使任意服务可以与客户端通信，还支持创建通信分组，实现消息组播。
- RPC支持（`Remote Procedure Call`）：支持服务、实体、客户端和分组间的RPC调用，基于`GAP协议`，支持可变类型，简单易用。支持单程通知RCP与有响应RPC。
- 日志（`Logger`）：基于Zap。
- 配置（`Config`）：基于Viper，支持本地本地配置与远端配置。
- 数据库（`DB`）：支持连接关系型数据库（基于`GORM`）、Redis、MongoDB。

## 目录
| Directory                                                                             | Description |
|---------------------------------------------------------------------------------------| ----------- |
| [/](https://github.com/pangdogs/framework)                                            | 开发应用时常用的类型与函数。|
| [/addins/broker](https://github.com/pangdogs/framework/tree/main/addins/broker)       | 消息队列中间件。|
| [/addins/conf](https://github.com/pangdogs/framework/tree/main/addins/conf)           | 配置系统。|
| [/addins/db](https://github.com/pangdogs/framework/tree/main/addins/db)               | 支持数据库。|
| [/addins/dentq](https://github.com/pangdogs/framework/tree/main/addins/dentq)         | 支持分布式实体查询。|
| [/addins/dentr](https://github.com/pangdogs/framework/tree/main/addins/dentr)         | 支持分布式实体注册。|
| [/addins/discovery](https://github.com/pangdogs/framework/tree/main/addins/discovery) | 服务发现。|
| [/addins/dsvc](https://github.com/pangdogs/framework/tree/main/addins/dsvc)           | 支持分布式服务。|
| [/addins/dsync](https://github.com/pangdogs/framework/tree/main/addins/dsync)         | 分布式锁。|
| [/addins/gate](https://github.com/pangdogs/framework/tree/main/addins/gate)           | 实现GTP网关。|
| [/addins/log](https://github.com/pangdogs/framework/tree/main/addins/log)             | 日志系统。|
| [/addins/router](https://github.com/pangdogs/framework/tree/main/addins/router)       | 客户端路由系统。|
| [/addins/rpc](https://github.com/pangdogs/framework/tree/main/addins/rpc)             | RPC系统。|
| [/addins/rpcstack](https://github.com/pangdogs/framework/tree/main/addins/rpcstack)   | 支持RPC堆栈。|
| [/net/gap](https://github.com/pangdogs/framework/tree/main/net/gap)                   | GAP协议实现。|
| [/net/gtp](https://github.com/pangdogs/framework/tree/main/net/gtp)                   | GTP协议实现。|
| [/net/netpath](https://github.com/pangdogs/framework/tree/main/net/netpath)           | 服务节点地址结构。|
| [/utils](https://github.com/pangdogs/framework/tree/main/utils)                       | 一些工具类与函数。 |

## 示例

详见： [Examples](https://github.com/pangdogs/examples)

## 安装
```
go get -u git.golaxy.org/framework
```
