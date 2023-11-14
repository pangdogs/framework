package distributed

import (
	"fmt"
	"kit.golaxy.org/golaxy"
	"time"
)

// Option 所有选项设置器
type Option struct{}

// DistributedOptions 所有选项
type DistributedOptions struct {
	RefreshInterval time.Duration // 服务刷新间隔
	FutureTimeout   time.Duration // 异步模型Future超时时间
}

// DistributedOption 选项设置器
type DistributedOption func(options *DistributedOptions)

// Default 默认值
func (Option) Default() DistributedOption {
	return func(options *DistributedOptions) {
		Option{}.RefreshInterval(3 * time.Second)(options)
		Option{}.FutureTimeout(5 * time.Second)(options)
	}
}

// RefreshInterval 刷新服务信息间隔
func (Option) RefreshInterval(d time.Duration) DistributedOption {
	return func(o *DistributedOptions) {
		if d <= 0 {
			panic(fmt.Errorf("%w: option RefreshInterval can't be set to a value less equal 0", golaxy.ErrArgs))
		}
		o.RefreshInterval = d
	}
}

// FutureTimeout 异步模型Future超时时间
func (Option) FutureTimeout(d time.Duration) DistributedOption {
	return func(options *DistributedOptions) {
		if d <= 0 {
			panic(fmt.Errorf("%w: option FutureTimeout can't be set to a value less equal 0", golaxy.ErrArgs))
		}
		options.FutureTimeout = d
	}
}
