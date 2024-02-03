// Package gap Golaxy应用层协议（golaxy application protocol），适用于开发应用层通信消息，需要工作在GTP协议或MQ之上，支持消息判重，解决了幂等性问题。
/*
	- 可以实现消息（Msg）接口新增自定义消息。
	- 支持可变类型（Variant），提供了一些常用的内置类型，也可以实现可变类型值（variant.Value）接口，新增自定义类型。
	- 支持自己实现消息（Msg）接口或可变类型值（variant.Value）接口，扩展支持protobuf（Protocol Buffers）等消息结构。
	- 支持消息判重，解决幂等性问题。
*/
package gap
