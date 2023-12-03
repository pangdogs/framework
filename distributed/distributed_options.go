package distributed

import (
	"fmt"
	"kit.golaxy.org/golaxy"
	"kit.golaxy.org/golaxy/util/generic"
	"kit.golaxy.org/golaxy/util/option"
	"kit.golaxy.org/plugins/gap"
	"time"
)

// Option 所有选项设置器
type Option struct{}

type (
	RecvMsgHandler = generic.DelegateFunc2[string, gap.MsgPacket, error] // 接收消息的处理器
)

// DistributedOptions 所有选项
type DistributedOptions struct {
	RefreshInterval   time.Duration   // 服务信息刷新间隔
	FutureTimeout     time.Duration   // 异步模型Future超时时间
	DecoderMsgCreator gap.IMsgCreator // 消息包解码器的消息构建器
	RecvMsgHandler    RecvMsgHandler  // 接收消息的处理器
}

// Default 默认值
func (Option) Default() option.Setting[DistributedOptions] {
	return func(options *DistributedOptions) {
		Option{}.RefreshInterval(3 * time.Second)(options)
		Option{}.FutureTimeout(5 * time.Second)(options)
		Option{}.DecoderMsgCreator(gap.DefaultMsgCreator())(options)
		Option{}.RecvMsgHandler(nil)(options)
	}
}

// RefreshInterval 服务信息刷新间隔
func (Option) RefreshInterval(d time.Duration) option.Setting[DistributedOptions] {
	return func(o *DistributedOptions) {
		if d <= 0 {
			panic(fmt.Errorf("%w: option RefreshInterval can't be set to a value less equal 0", golaxy.ErrArgs))
		}
		o.RefreshInterval = d
	}
}

// FutureTimeout 异步模型Future超时时间
func (Option) FutureTimeout(d time.Duration) option.Setting[DistributedOptions] {
	return func(options *DistributedOptions) {
		if d <= 0 {
			panic(fmt.Errorf("%w: option FutureTimeout can't be set to a value less equal 0", golaxy.ErrArgs))
		}
		options.FutureTimeout = d
	}
}

func (Option) DecoderMsgCreator(mc gap.IMsgCreator) option.Setting[DistributedOptions] {
	return func(options *DistributedOptions) {
		if mc == nil {
			panic(fmt.Errorf("%w: option DecoderMsgCreator can't be assigned to nil", golaxy.ErrArgs))
		}
		options.DecoderMsgCreator = mc
	}
}

func (Option) RecvMsgHandler(handler RecvMsgHandler) option.Setting[DistributedOptions] {
	return func(options *DistributedOptions) {
		options.RecvMsgHandler = handler
	}
}
