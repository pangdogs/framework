package distributed

import (
	"fmt"
	"kit.golaxy.org/golaxy"
	"time"
)

type Option struct{}

type DistributedOptions struct {
	RefreshInterval time.Duration // 服务刷新间隔
	FutureTimeout   time.Duration // 异步模型Future超时时间
}

type DistributedOption func(options *DistributedOptions)

func (Option) Default() DistributedOption {
	return func(options *DistributedOptions) {
		Option{}.RefreshInterval(3 * time.Second)(options)
		Option{}.FutureTimeout(5 * time.Second)(options)
	}
}

func (Option) RefreshInterval(d time.Duration) DistributedOption {
	return func(o *DistributedOptions) {
		if d <= 0 {
			panic(fmt.Errorf("%w: option RefreshInterval can't be set to a value less equal 0", golaxy.ErrArgs))
		}
		o.RefreshInterval = d
	}
}

func (Option) FutureTimeout(d time.Duration) DistributedOption {
	return func(options *DistributedOptions) {
		if d <= 0 {
			panic(fmt.Errorf("%w: option FutureTimeout can't be set to a value less equal 0", golaxy.ErrArgs))
		}
		options.FutureTimeout = d
	}
}
