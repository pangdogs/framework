# Framework
[English](./README.md) | [简体中文](./README.zh_CN.md)

## Overview
[**Golaxy Distributed Service Development Framework**](https://github.com/pangdogs/framework) is the service-side companion to [**Golaxy Core**](https://github.com/pangdogs/core). It builds a distributed service runtime on top of the EC system and actor-thread model, and packages the infrastructure commonly needed by real-time systems such as game servers, gateways, and remote control platforms.

This repository is organized around three layers:

- `framework`: application bootstrap, service/runtime assembly, lifecycle hooks, and entity/component helpers.
- `addins`: reusable infrastructure integrations such as broker, discovery, router, RPC, gateway, and database access.
- `net`: the protocol stack used by the communication layer, including GTP transport and GAP application messages.

## Architecture
### Core concepts
- `App` collects service assemblers, loads CLI/configuration with Cobra and Viper, and starts the requested service replicas.
- `IService` wraps a service context and exposes add-ins such as broker, discovery, distributed sync, distributed service, distributed entity query, and RPC.
- `IRuntime` wraps a runtime context, installs runtime add-ins, and owns entity execution.
- `EntityBehavior` and `ComponentBehavior` provide typed access back to the owning runtime and service.
- `BuildRuntime`, `BuildEntityPT`, and `BuildEntity` are the main builder APIs for runtime and entity creation.

### Default add-ins
Unless a service or runtime installs its own implementation during lifecycle hooks, the framework assembles these add-ins automatically.

Service-level defaults:

- `log`
- `conf`
- `broker` via NATS
- `discovery` via ETCD
- `dsync` via ETCD
- `dsvc`
- `dent` querier
- `rpc`

Runtime-level defaults:

- `log`
- `rpcstack`
- `dent` registry

## Features
- Bootstrap stateful or stateless services with configurable replica counts.
- Load configuration from flags, environment variables, local files, or Viper remote providers.
- Discover services, advertise nodes, and issue future-based inter-service calls.
- Register and query distributed entities across runtimes and services.
- Use a built-in broker layer, distributed synchronization, and SQL/Redis/MongoDB integrations.
- Serve long-lived TCP/WebSocket connections with the GTP transport stack.
- Exchange routed messages and RPC requests/replies with GAP messages and variant payloads.
- Build gateways, routers, service RPC processors, and RPC clients from shared protocol components.

## Minimal bootstrap
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

The service name passed to `SetAssembler` is used as:

- the configuration subtree returned by `svc.ServiceConf()`
- the default key inside `startup.services`
- the logical service name announced by distributed-service related add-ins

## Configuration and startup
`App` registers a common set of flags before boot:

| Flag | Purpose |
| --- | --- |
| `log.*` | Logger level, encoder, format, and async buffering settings |
| `conf.*` | Environment prefix, local config path, and remote config provider settings |
| `nats.*` | Default broker connection settings |
| `etcd.*` | Default ETCD connection settings |
| `service.*` | Version, metadata, keepalive TTLs, future timeout, entity TTL, and panic recovery |
| `startup.services` | Replica count per registered service name |
| `pprof.*` | Optional pprof listener configuration |

Typical startup command:

```bash
your-app \
  --startup.services lobby=2 \
  --nats.address localhost:4222 \
  --etcd.address localhost:2379 \
  --conf.local_path ./config.yaml
```

Service-specific settings should live under the service name:

```yaml
lobby:
  tick_interval: 50ms
  gate:
    tcp_address: 0.0.0.0:7001
```

Inside service code, `svc.ServiceConf()` resolves to the `lobby` subtree while `svc.AppConf()` still exposes the full merged configuration.

## Package layout
| Path | Responsibility |
| --- | --- |
| [`./`](./) | App bootstrap, service/runtime/entity builders, lifecycle contracts, async helpers |
| [`./addins`](./addins) | Convenience re-exports for built-in add-in installers and option helpers |
| [`./addins/broker`](./addins/broker) | Broker abstraction and NATS implementation |
| [`./addins/conf`](./addins/conf) | Viper-backed configuration add-in |
| [`./addins/db`](./addins/db) | DB injection and integrations for SQL, Redis, and MongoDB |
| [`./addins/dent`](./addins/dent) | Distributed entity query and registry add-ins |
| [`./addins/discovery`](./addins/discovery) | Service discovery abstraction and ETCD implementation |
| [`./addins/dsvc`](./addins/dsvc) | Distributed service addressing and future-based service calls |
| [`./addins/dsync`](./addins/dsync) | Distributed synchronization with ETCD and Redis implementations |
| [`./addins/gate`](./addins/gate) | Gateway/session management on top of GTP |
| [`./addins/gate/cli`](./addins/gate/cli) | Low-level client for GTP/GAP endpoints |
| [`./addins/log`](./addins/log) | Zap-backed logging add-in |
| [`./addins/router`](./addins/router) | Session routing, mapping, groups, and multicast helpers |
| [`./addins/rpc`](./addins/rpc) | RPC facade, call-path helpers, processors, and RPC client tooling |
| [`./addins/rpcstack`](./addins/rpcstack) | RPC call chain/context stack |
| [`./net/gap`](./net/gap) | GAP messages, codec, and dynamic variant values |
| [`./net/gtp`](./net/gtp) | GTP messages, codec, crypto/compression methods, and transport machinery |
| [`./net/netpath`](./net/netpath) | Helpers for distributed service/node path formats |
| [`./utils`](./utils) | Binary and concurrency helpers used by the framework internals |

## Protocol stack
- `net/gtp` is the transport layer. It defines handshake messages, cipher negotiation, authentication, compression, and reliable packet delivery over TCP or WebSocket.
- `net/gap` sits above GTP or broker-delivered payloads. It carries forwarded messages, RPC requests/replies, one-way RPC, and extensible application payloads.
- `addins/gate`, `addins/router`, and `addins/rpc` build service-facing communication features on top of those two layers.

## Requirements
- Go `1.25+`
- NATS for the default broker add-in
- ETCD for the default discovery, distributed sync, and distributed entity registry/query add-ins
- Optional Redis if you choose Redis-backed distributed sync or Redis DB access
- Optional MongoDB or SQL databases depending on the enabled DB add-ins

## Installation
```bash
go get -u git.golaxy.org/framework
```

## Examples
See [pangdogs/examples](https://github.com/pangdogs/examples) for end-to-end services, gateways, and RPC usage.

## Related repositories
- [Golaxy Distributed Service Development Framework Core](https://github.com/pangdogs/core)
- [Golaxy Game Server Scaffold](https://github.com/pangdogs/scaffold)
