package dserv

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/net/gap"
	"time"
)

// Option 所有选项设置器
type Option struct{}

type (
	RecvMsgHandler = generic.DelegateFunc2[string, gap.MsgPacket, error] // 接收消息的处理器
)

// DistServiceOptions 所有选项
type DistServiceOptions struct {
	Version           string            // 服务版本号
	Meta              map[string]string // 服务元数据，以键值对的形式保存附加信息
	Domain            string            // 服务地址域
	TTL               time.Duration     // 服务信息过期时间
	FutureTimeout     time.Duration     // 异步模型Future超时时间
	DecoderMsgCreator gap.IMsgCreator   // 消息包解码器的消息构建器
	RecvMsgHandler    RecvMsgHandler    // 接收消息的处理器（优先级低于监控器）
}

// Default 默认值
func (Option) Default() option.Setting[DistServiceOptions] {
	return func(options *DistServiceOptions) {
		Option{}.Version("")(options)
		Option{}.Meta(nil)(options)
		Option{}.Domain("service")(options)
		Option{}.TTL(10 * time.Second)(options)
		Option{}.FutureTimeout(5 * time.Second)(options)
		Option{}.DecoderMsgCreator(gap.DefaultMsgCreator())(options)
		Option{}.RecvMsgHandler(nil)(options)
	}
}

// Version 服务版本号
func (Option) Version(version string) option.Setting[DistServiceOptions] {
	return func(o *DistServiceOptions) {
		o.Version = version
	}
}

// Meta 服务元数据，以键值对的形式保存附加信息
func (Option) Meta(meta map[string]string) option.Setting[DistServiceOptions] {
	return func(o *DistServiceOptions) {
		o.Meta = meta
	}
}

// Domain 服务地址域
func (Option) Domain(domain string) option.Setting[DistServiceOptions] {
	return func(o *DistServiceOptions) {
		o.Domain = domain
	}
}

// TTL 服务信息过期时间
func (Option) TTL(ttl time.Duration) option.Setting[DistServiceOptions] {
	return func(o *DistServiceOptions) {
		if ttl < 3*time.Second {
			panic(fmt.Errorf("%w: option TTL can't be set to a value less than 3 second", core.ErrArgs))
		}
		o.TTL = ttl
	}
}

// FutureTimeout 异步模型Future超时时间
func (Option) FutureTimeout(d time.Duration) option.Setting[DistServiceOptions] {
	return func(options *DistServiceOptions) {
		if d <= 0 {
			panic(fmt.Errorf("%w: option FutureTimeout can't be set to a value less equal 0", core.ErrArgs))
		}
		options.FutureTimeout = d
	}
}

// DecoderMsgCreator 消息包解码器的消息构建器
func (Option) DecoderMsgCreator(mc gap.IMsgCreator) option.Setting[DistServiceOptions] {
	return func(options *DistServiceOptions) {
		if mc == nil {
			panic(fmt.Errorf("%w: option DecoderMsgCreator can't be assigned to nil", core.ErrArgs))
		}
		options.DecoderMsgCreator = mc
	}
}

// RecvMsgHandler 接收消息的处理器
func (Option) RecvMsgHandler(handler RecvMsgHandler) option.Setting[DistServiceOptions] {
	return func(options *DistServiceOptions) {
		options.RecvMsgHandler = handler
	}
}
