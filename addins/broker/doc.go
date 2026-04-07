// Package broker 定义框架服务使用的发布订阅 broker 抽象。
//
// 可通过 AddIn 从 service context 中获取 IBroker，并通过 broker_nats
// 之类的具体实现提供底层传输能力。
package broker
