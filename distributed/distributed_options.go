package distributed

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/plugins/gap"
	"git.golaxy.org/plugins/registry"
	"time"
)

// Option 所有选项设置器
type Option struct{}

type (
	RecvMsgHandler = generic.DelegateFunc2[string, gap.MsgPacket, error] // 接收消息的处理器
)

// DistributedOptions 所有选项
type DistributedOptions struct {
	Version           string              // 服务版本号
	Metadata          map[string]string   // 服务元数据，以键值对的形式保存附加信息
	Endpoints         []registry.Endpoint // 服务端点列表
	Domain            string              // 服务地址域
	RefreshInterval   time.Duration       // 服务信息刷新间隔
	FutureTimeout     time.Duration       // 异步模型Future超时时间
	DecoderMsgCreator gap.IMsgCreator     // 消息包解码器的消息构建器
	RecvMsgHandler    RecvMsgHandler      // 接收消息的处理器（优先级低于监控器）
}

// Default 默认值
func (Option) Default() option.Setting[DistributedOptions] {
	return func(options *DistributedOptions) {
		Option{}.Version("")(options)
		Option{}.Metadata(nil)(options)
		Option{}.Endpoints(nil)(options)
		Option{}.Domain("service")(options)
		Option{}.RefreshInterval(3 * time.Second)(options)
		Option{}.FutureTimeout(5 * time.Second)(options)
		Option{}.DecoderMsgCreator(gap.DefaultMsgCreator())(options)
		Option{}.RecvMsgHandler(nil)(options)
	}
}

// Version 服务版本号
func (Option) Version(version string) option.Setting[DistributedOptions] {
	return func(o *DistributedOptions) {
		o.Version = version
	}
}

// Metadata 服务元数据，以键值对的形式保存附加信息
func (Option) Metadata(meta map[string]string) option.Setting[DistributedOptions] {
	return func(o *DistributedOptions) {
		o.Metadata = meta
	}
}

// Endpoints 服务端点列表
func (Option) Endpoints(endpoints []registry.Endpoint) option.Setting[DistributedOptions] {
	return func(o *DistributedOptions) {
		o.Endpoints = endpoints
	}
}

// Domain 服务地址域
func (Option) Domain(domain string) option.Setting[DistributedOptions] {
	return func(o *DistributedOptions) {
		o.Domain = domain
	}
}

// RefreshInterval 服务信息刷新间隔
func (Option) RefreshInterval(d time.Duration) option.Setting[DistributedOptions] {
	return func(o *DistributedOptions) {
		if d <= 0 {
			panic(fmt.Errorf("%w: option RefreshInterval can't be set to a value less equal 0", core.ErrArgs))
		}
		o.RefreshInterval = d
	}
}

// FutureTimeout 异步模型Future超时时间
func (Option) FutureTimeout(d time.Duration) option.Setting[DistributedOptions] {
	return func(options *DistributedOptions) {
		if d <= 0 {
			panic(fmt.Errorf("%w: option FutureTimeout can't be set to a value less equal 0", core.ErrArgs))
		}
		options.FutureTimeout = d
	}
}

// DecoderMsgCreator 消息包解码器的消息构建器
func (Option) DecoderMsgCreator(mc gap.IMsgCreator) option.Setting[DistributedOptions] {
	return func(options *DistributedOptions) {
		if mc == nil {
			panic(fmt.Errorf("%w: option DecoderMsgCreator can't be assigned to nil", core.ErrArgs))
		}
		options.DecoderMsgCreator = mc
	}
}

// RecvMsgHandler 接收消息的处理器
func (Option) RecvMsgHandler(handler RecvMsgHandler) option.Setting[DistributedOptions] {
	return func(options *DistributedOptions) {
		options.RecvMsgHandler = handler
	}
}
