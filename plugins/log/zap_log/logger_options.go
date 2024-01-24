package zap_log

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/util/option"
	"go.uber.org/zap"
)

// Option 所有选项设置器
type Option struct{}

// LoggerOptions 所有选项
type LoggerOptions struct {
	ZapLogger   *zap.Logger
	ServiceInfo bool
	RuntimeInfo bool
	CallerSkip  int
}

// Default 默认值
func (Option) Default() option.Setting[LoggerOptions] {
	return func(options *LoggerOptions) {
		Option{}.ZapLogger(zap.NewExample())(options)
		Option{}.ServiceInfo(false)(options)
		Option{}.RuntimeInfo(false)(options)
		Option{}.CallerSkip(2)(options)
	}
}

// ZapLogger zap logger
func (Option) ZapLogger(logger *zap.Logger) option.Setting[LoggerOptions] {
	return func(options *LoggerOptions) {
		if logger == nil {
			panic(fmt.Errorf("%w: option ZapLogger can't be assigned to nil", core.ErrArgs))
		}
		options.ZapLogger = logger
	}
}

// ServiceInfo 添加service信息
func (Option) ServiceInfo(b bool) option.Setting[LoggerOptions] {
	return func(options *LoggerOptions) {
		options.ServiceInfo = b
	}
}

// RuntimeInfo 添加runtime信息
func (Option) RuntimeInfo(b bool) option.Setting[LoggerOptions] {
	return func(options *LoggerOptions) {
		options.RuntimeInfo = b
	}
}

// CallerSkip 调用堆栈skip值，用于打印调用堆栈信息
func (Option) CallerSkip(skip int) option.Setting[LoggerOptions] {
	return func(options *LoggerOptions) {
		if skip < 0 {
			panic(fmt.Errorf("%w: option CallerSkip can't be set to a value less than 0", core.ErrArgs))
		}
		options.CallerSkip = skip
	}
}
