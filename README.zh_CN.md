# Framework
[English](./README.md) | [简体中文](./README.zh_CN.md)

## 简介
[**Golaxy 分布式服务开发框架**](https://github.com/pangdogs/framework) 是 [**Golaxy Core**](https://github.com/pangdogs/core) 的服务端扩展层。它建立在 EC 系统和 Actor 线程模型之上，把实时通信系统常见的基础设施能力封装为统一的服务框架，适合游戏服务端、网关、远程控制平台等场景。

仓库当前主要分为三层：

- `framework`：应用启动、服务/运行时装配、生命周期接口，以及实体/组件构建辅助。
- `addins`：可复用的基础设施插件，例如 broker、服务发现、路由、RPC、网关和数据库访问。
- `net`：通信层协议栈，包括 GTP 传输协议和 GAP 应用层消息协议。

## 架构说明
### 核心概念
- `App` 负责收集服务装配器，使用 Cobra/Viper 加载命令行与配置，并按配置启动多个服务副本。
- `IService` 封装服务上下文，统一暴露 broker、服务发现、分布式同步、分布式服务、分布式实体查询和 RPC 等能力。
- `IRuntime` 封装运行时上下文，负责安装运行时级 add-in，并承载实体执行。
- `EntityBehavior` 和 `ComponentBehavior` 为实体、组件提供回到所属 runtime 和 service 的强类型访问。
- `BuildRuntime`、`BuildEntityPT`、`BuildEntity` 是创建运行时、实体原型和实体实例的主要入口。

### 默认装配的 add-in
如果业务代码没有在生命周期阶段安装自定义实现，框架会自动装配以下 add-in。

服务级默认 add-in：

- `log`
- `conf`
- `broker`，默认实现为 NATS
- `discovery`，默认实现为 ETCD
- `dsync`，默认实现为 ETCD
- `dsvc`
- `dent` 查询端
- `rpc`

运行时级默认 add-in：

- `log`
- `rpcstack`
- `dent` 注册端

## 功能特性
- 支持有状态或无状态服务，并支持按副本数启动多个实例。
- 支持从命令行、环境变量、本地配置文件、Viper 远程配置源加载配置。
- 支持服务注册发现、节点发布，以及基于 `Future` 的跨服务调用。
- 支持分布式实体注册、查询，以及 runtime 内实体创建。
- 内置 broker 层、分布式同步能力，以及 SQL/Redis/MongoDB 集成。
- 提供基于 TCP/WebSocket 的 GTP 长连接传输层。
- 提供 GAP 应用层消息、转发、RPC 请求/响应与动态变体类型。
- 可在统一协议栈上构建网关、路由、RPC 处理器和 RPC 客户端。

## 最小启动示例
```go
package main

import "git.golaxy.org/framework"

type LobbyService struct {
	framework.ServiceBehavior
}

func (svc *LobbyService) OnStarted(s framework.IService) {
	s.BuildRuntime().
		SetName("main").
		SetEnableFrame(true).
		SetFPS(20).
		New()
}

func main() {
	framework.NewApp().
		SetAssembler("lobby", &LobbyService{}).
		Run()
}
```

传给 `SetAssembler` 的服务名会同时用于：

- `svc.ServiceConf()` 返回的服务配置子树名
- `startup.services` 中的默认服务键名
- 分布式服务相关 add-in 对外发布的逻辑服务名

## 配置与启动
`App` 在启动前会统一注册以下配置项：

| 配置项 | 作用 |
| --- | --- |
| `log.*` | 日志级别、编码器、输出格式和异步缓冲参数 |
| `conf.*` | 环境变量前缀、本地配置路径、远程配置源参数 |
| `nats.*` | 默认 broker 连接参数 |
| `etcd.*` | 默认 ETCD 连接参数 |
| `service.*` | 服务版本、元数据、保活 TTL、Future 超时、实体 TTL、panic 自动恢复 |
| `startup.services` | 每个服务名对应的启动副本数 |
| `pprof.*` | 可选的 pprof 监听参数 |

典型启动命令：

```bash
your-app \
  --startup.services lobby=2 \
  --nats.address localhost:4222 \
  --etcd.address localhost:2379 \
  --conf.local_path ./config.yaml
```

服务专属配置建议放在服务名子树下：

```yaml
lobby:
  tick_interval: 50ms
  gate:
    tcp_address: 0.0.0.0:7001
```

在代码中，`svc.ServiceConf()` 会定位到 `lobby` 子树，而 `svc.AppConf()` 仍然返回完整的合并后配置。

## 目录说明
| 路径 | 职责 |
| --- | --- |
| [`./`](./) | App 启动、服务/运行时/实体构建器、生命周期接口、异步辅助 |
| [`./addins`](./addins) | 内置 add-in 安装入口与常用 option 辅助的聚合导出 |
| [`./addins/broker`](./addins/broker) | Broker 抽象与 NATS 实现 |
| [`./addins/conf`](./addins/conf) | 基于 Viper 的配置 add-in |
| [`./addins/db`](./addins/db) | SQL、Redis、MongoDB 集成与数据库注入辅助 |
| [`./addins/dent`](./addins/dent) | 分布式实体查询与注册 add-in |
| [`./addins/discovery`](./addins/discovery) | 服务发现抽象与 ETCD 实现 |
| [`./addins/dsvc`](./addins/dsvc) | 分布式服务寻址与基于 Future 的服务调用 |
| [`./addins/dsync`](./addins/dsync) | 分布式同步能力及 ETCD/Redis 实现 |
| [`./addins/gate`](./addins/gate) | 构建在 GTP 之上的网关与会话管理 |
| [`./addins/gate/cli`](./addins/gate/cli) | 面向 GTP/GAP 端点的底层客户端 |
| [`./addins/log`](./addins/log) | 基于 Zap 的日志 add-in |
| [`./addins/router`](./addins/router) | 会话路由、映射、分组和组播辅助 |
| [`./addins/rpc`](./addins/rpc) | RPC 门面、调用路径、处理器和 RPC 客户端工具 |
| [`./addins/rpcstack`](./addins/rpcstack) | RPC 调用链上下文栈 |
| [`./net/gap`](./net/gap) | GAP 消息、编解码和动态变体类型 |
| [`./net/gtp`](./net/gtp) | GTP 消息、编解码、加密/压缩算法和传输实现 |
| [`./net/netpath`](./net/netpath) | 分布式服务/节点路径格式辅助 |
| [`./utils`](./utils) | 框架内部使用的二进制与并发辅助工具 |

## 协议栈分层
- `net/gtp` 是传输层，定义握手消息、密码套件协商、链路鉴权、压缩，以及基于 TCP/WebSocket 的可靠消息传输。
- `net/gap` 位于 GTP 或 MQ 之上，用于承载转发消息、RPC 请求/响应、单向 RPC 和可扩展应用层负载。
- `addins/gate`、`addins/router`、`addins/rpc` 在这两层协议之上构建面向业务的通信能力。

## 环境要求
- Go `1.25+`
- 默认 broker 依赖 NATS
- 默认服务发现、分布式同步、分布式实体注册/查询依赖 ETCD
- 如果选择 Redis 版分布式同步或 Redis 数据库 add-in，则需要 Redis
- 根据启用的数据库 add-in，可选 MongoDB 或各类 SQL 数据库

## 安装
```bash
go get -u git.golaxy.org/framework
```

## 示例
完整示例可参考 [pangdogs/examples](https://github.com/pangdogs/examples)。

## 相关仓库
- [Golaxy Distributed Service Development Framework Core](https://github.com/pangdogs/core)
- [Golaxy 游戏服务脚手架](https://github.com/pangdogs/scaffold)
