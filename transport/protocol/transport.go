package tcp

import (
	"kit.golaxy.org/plugins/transport/codec"
	"net"
)

// Trans 传输协议
type Trans struct {
	Conn       net.Conn       // 网络连接
	Encoder    codec.IEncoder // 消息包编码器
	Decoder    codec.IDecoder // 消息包解码器
	RetryTimes int            // io超时重试次数
}
