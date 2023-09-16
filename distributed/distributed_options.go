package distributed

import (
	"time"
)

type Option struct{}

type DistributedOptions struct {
	RefreshInterval time.Duration
}

type DistributedOption func(options *DistributedOptions)

func (Option) Default() DistributedOption {
	return func(options *DistributedOptions) {
		Option{}.RefreshInterval(3 * time.Second)(options)
	}
}

func (Option) RefreshInterval(d time.Duration) DistributedOption {
	return func(o *DistributedOptions) {
		if d <= 0 {
			panic("option RefreshInterval can't be set to a value less equal 0")
		}
		o.RefreshInterval = d
	}
}
