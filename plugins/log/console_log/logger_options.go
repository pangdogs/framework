package console_log

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/plugins/log"
	"time"
)

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

var With _Option

type _Option struct{}

// Default 默认值
func (_Option) Default() option.Setting[LoggerOptions] {
	return func(options *LoggerOptions) {
		With.Level(log.InfoLevel)(options)
		With.Development(false)
		With.ServiceInfo(false)(options)
		With.RuntimeInfo(false)(options)
		With.Separator(`|`)(options)
		With.TimestampLayout(time.RFC3339Nano)(options)
		With.CallerFullName(false)(options)
		With.CallerSkip(3)(options)
	}
}

// Level 日志等级
func (_Option) Level(level log.Level) option.Setting[LoggerOptions] {
	return func(options *LoggerOptions) {
		options.Level = level
	}
}

// Development 开发模式
func (_Option) Development(b bool) option.Setting[LoggerOptions] {
	return func(options *LoggerOptions) {
		options.Development = b
	}
}

// ServiceInfo 添加service信息
func (_Option) ServiceInfo(b bool) option.Setting[LoggerOptions] {
	return func(options *LoggerOptions) {
		options.ServiceInfo = b
	}
}

// RuntimeInfo 添加runtime信息
func (_Option) RuntimeInfo(b bool) option.Setting[LoggerOptions] {
	return func(options *LoggerOptions) {
		options.RuntimeInfo = b
	}
}

// Separator 分隔符
func (_Option) Separator(sp string) option.Setting[LoggerOptions] {
	return func(options *LoggerOptions) {
		options.Separator = sp
	}
}

// TimestampLayout 时间格式
func (_Option) TimestampLayout(layout string) option.Setting[LoggerOptions] {
	return func(options *LoggerOptions) {
		options.TimestampLayout = layout
	}
}

// CallerFullName 是否打印完整调用堆栈信息
func (_Option) CallerFullName(b bool) option.Setting[LoggerOptions] {
	return func(options *LoggerOptions) {
		options.CallerFullName = b
	}
}

// CallerSkip 调用堆栈skip值，用于打印调用堆栈信息
func (_Option) CallerSkip(skip int) option.Setting[LoggerOptions] {
	return func(options *LoggerOptions) {
		if skip < 0 {
			panic(fmt.Errorf("%w: option CallerSkip can't be set to a value less than 0", core.ErrArgs))
		}
		options.CallerSkip = skip
	}
}
