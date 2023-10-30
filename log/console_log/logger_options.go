package console_log

import (
	"fmt"
	"kit.golaxy.org/golaxy"
	"kit.golaxy.org/plugins/log"
	"time"
)

// Option 所有选项设置器
type Option struct{}

// LoggerOptions 所有选项
type LoggerOptions struct {
	Level           log.Level
	Development     bool
	ServiceInfo     bool
	RuntimeInfo     bool
	Separator       string
	TimestampLayout string
	CallerFullName  bool
	CallerSkip      int
}

// LoggerOption 选项设置器
type LoggerOption func(options *LoggerOptions)

// Default 默认值
func (Option) Default() LoggerOption {
	return func(options *LoggerOptions) {
		Option{}.Level(log.InfoLevel)(options)
		Option{}.Development(false)
		Option{}.ServiceInfo(true)(options)
		Option{}.RuntimeInfo(true)(options)
		Option{}.Separator(`|`)(options)
		Option{}.TimestampLayout(time.RFC3339Nano)(options)
		Option{}.CallerFullName(false)(options)
		Option{}.CallerSkip(2)(options)
	}
}

// Level 日志等级
func (Option) Level(level log.Level) LoggerOption {
	return func(options *LoggerOptions) {
		options.Level = level
	}
}

// Development 开发模式
func (Option) Development(b bool) LoggerOption {
	return func(options *LoggerOptions) {
		options.Development = b
	}
}

// ServiceInfo 添加service信息
func (Option) ServiceInfo(b bool) LoggerOption {
	return func(options *LoggerOptions) {
		options.ServiceInfo = b
	}
}

// RuntimeInfo 添加runtime信息
func (Option) RuntimeInfo(b bool) LoggerOption {
	return func(options *LoggerOptions) {
		options.RuntimeInfo = b
	}
}

// Separator 分隔符
func (Option) Separator(sp string) LoggerOption {
	return func(options *LoggerOptions) {
		options.Separator = sp
	}
}

// TimestampLayout 时间格式
func (Option) TimestampLayout(layout string) LoggerOption {
	return func(options *LoggerOptions) {
		options.TimestampLayout = layout
	}
}

// CallerFullName 是否打印完整调用堆栈信息
func (Option) CallerFullName(b bool) LoggerOption {
	return func(options *LoggerOptions) {
		options.CallerFullName = b
	}
}

// CallerSkip 调用堆栈skip值，用于打印调用堆栈信息
func (Option) CallerSkip(skip int) LoggerOption {
	return func(options *LoggerOptions) {
		if skip < 0 {
			panic(fmt.Errorf("%w: option CallerSkip can't be set to a value less than 0", golaxy.ErrArgs))
		}
		options.CallerSkip = skip
	}
}
