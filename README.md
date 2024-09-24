# Framework
[English](./README.md) | [简体中文](./README.zh_CN.md)

## Introduction
[**Golaxy Distributed Service Development Framework**](https://github.com/pangdogs/framework) aims to provide a comprehensive server-side solution for real-time communication applications. Based on the [**core**](https://github.com/pangdogs/core) of the EC system and Actor thread model, the framework implements all dependency functions for distributed services. It is designed to be simple and easy to use, making it particularly suitable for developing games and remote control systems.

## Features
The framework supports the development of stateful (`Stateful`) or stateless (`Stateless`) distributed services with the following features:

- **MQ and Broker**: Based on NATS, supports message queue and event-driven architecture.
- **Service Discovery**: Based on ETCD, also supports Redis (for demo purposes due to lack of data version control).
- **Distributed Synchronization**: Supports distributed locking with ETCD or Redis, default is ETCD.
- **Distributed Service**: Defines distributed service node address format, provides asynchronous model futures (`Future`), supports inter-service communication and horizontal scaling.
- **Distributed Entities**: Provides registration and query functions for distributed entities, supports communication between them.
- **GTP Protocol**: For long connections and real-time communication, works on reliable protocols (`TCP/WebSocket`), supports bi-directional signature verification, link encryption, link authentication, reconnect and retransmission, custom messages.
- **GAP Protocol**: For application layer communication messages, works on `GTP Protocol` or `MQ`, supports message deduplication, custom messages, custom variable types.
- **GTP Gate and Client**: Gateway and client based on `GTP Protocol`, supports `TCP/WebSocket` long connections.
- **Router**: Supports communication routing, session to entity mapping, client-service communication, communication grouping, and multicast messages.
- **RPC**: Supports RPC calls between services, entities, clients, and groups based on `GAP Protocol`, supports variable types, simple and easy to use. Supports one-way notification RPC and response RPC.
- **Logger**: Based on Zap.
- **Config**: Based on Viper, supports local and remote configurations.
- **Database**: Supports connection to relational databases (based on `GORM`), Redis, MongoDB.

## Directory
| Directory                                                                               | Description |
|-----------------------------------------------------------------------------------------| ----------- |
| [/](https://github.com/pangdogs/framework)                                              | Common types and functions for application development. |
| [/net/gap](https://github.com/pangdogs/framework/tree/main/net/gap)                     | GAP protocol implementation. |
| [/net/gtp](https://github.com/pangdogs/framework/tree/main/net/gtp)                     | GTP protocol implementation. |
| [/net/netpath](https://github.com/pangdogs/framework/tree/main/net/netpath)             | Service node address structure. |
| [/plugins/broker](https://github.com/pangdogs/framework/tree/main/plugins/broker)       | Message queue middleware. |
| [/plugins/conf](https://github.com/pangdogs/framework/tree/main/plugins/conf)           | Configuration system. |
| [/plugins/db](https://github.com/pangdogs/framework/tree/main/plugins/db)               | Database support. |
| [/plugins/dentq](https://github.com/pangdogs/framework/tree/main/plugins/dentq)         | Distributed entity query support. |
| [/plugins/dentr](https://github.com/pangdogs/framework/tree/main/plugins/dentr)         | Distributed entity registration support. |
| [/plugins/discovery](https://github.com/pangdogs/framework/tree/main/plugins/discovery) | Service discovery. |
| [/plugins/dsvc](https://github.com/pangdogs/framework/tree/main/plugins/dsvc)          | Distributed service support. |
| [/plugins/dsync](https://github.com/pangdogs/framework/tree/main/plugins/dsync)         | Distributed locking. |
| [/plugins/gate](https://github.com/pangdogs/framework/tree/main/plugins/gate)           | GTP gateway implementation. |
| [/plugins/log](https://github.com/pangdogs/framework/tree/main/plugins/log)             | Logging system. |
| [/plugins/router](https://github.com/pangdogs/framework/tree/main/plugins/router)       | Client routing system. |
| [/plugins/rpc](https://github.com/pangdogs/framework/tree/main/plugins/rpc)             | RPC system. |
| [/plugins/rpcstack](https://github.com/pangdogs/framework/tree/main/plugins/rpcstack)   | RPC stack support. |

## Examples
See: [Examples](https://github.com/pangdogs/examples)

## Installation
```bash
go get -u git.golaxy.org/framework
```
