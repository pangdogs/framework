package zap_log

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/option"
	"go.uber.org/zap"
)

// LoggerOptions 所有选项
type LoggerOptions struct {
	ZapLogger   *zap.Logger
	ServiceInfo bool
	RuntimeInfo bool
	CallerSkip  int
}

var With _Option

type _Option struct{}

// Default 默认值
func (_Option) Default() option.Setting[LoggerOptions] {
	return func(options *LoggerOptions) {
		With.ZapLogger(zap.NewExample())(options)
		With.ServiceInfo(false)(options)
		With.RuntimeInfo(false)(options)
		With.CallerSkip(2)(options)
	}
}

// ZapLogger zap logger
func (_Option) ZapLogger(logger *zap.Logger) option.Setting[LoggerOptions] {
	return func(options *LoggerOptions) {
		if logger == nil {
			panic(fmt.Errorf("%w: option ZapLogger can't be assigned to nil", core.ErrArgs))
		}
		options.ZapLogger = logger
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

// CallerSkip 调用堆栈skip值，用于打印调用堆栈信息
func (_Option) CallerSkip(skip int) option.Setting[LoggerOptions] {
	return func(options *LoggerOptions) {
		if skip < 0 {
			panic(fmt.Errorf("%w: option CallerSkip can't be set to a value less than 0", core.ErrArgs))
		}
		options.CallerSkip = skip
	}
}
