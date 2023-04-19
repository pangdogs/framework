package zap

import (
	"go.uber.org/zap"
)

type ZapOptions struct {
	ZapLogger *zap.Logger
}

type ZapOption func(options *ZapOptions)

type WithZapOption struct{}

func (WithZapOption) Default() ZapOption {
	return func(options *ZapOptions) {
		WithZapOption{}.ZapLogger(nil)(options)
	}
}

func (WithZapOption) ZapLogger(v *zap.Logger) ZapOption {
	return func(options *ZapOptions) {
		options.ZapLogger = v
	}
}

//
//var log *zap.SugaredLogger
//var atomicLevel zap.AtomicLevel
//
//func init() {
//	// 日志分割
//	hook := lumberjack.Logger{
//		Filename: "./log/florahub.log", // 日志文件路径
//		MaxSize:  100,                  // 每个日志文件保存100M
//	}
//	write := zapcore.AddSync(&hook)
//	// 设置日志级别
//	// debug 可以打印出 info debug warn
//	// info  级别可以打印 warn info
//	// warn  只能打印 warn
//	// debug->info->warn->error
//	level := zap.InfoLevel
//
//	encoderConfig := zapcore.EncoderConfig{
//		TimeKey:          "time",
//		LevelKey:         "level",
//		NameKey:          "logger",
//		CallerKey:        "linenum",
//		MessageKey:       "msg",
//		StacktraceKey:    "stacktrace",
//		LineEnding:       zapcore.DefaultLineEnding,
//		EncodeLevel:      zapcore.LowercaseLevelEncoder, // 小写编码器
//		EncodeTime:       zapcore.ISO8601TimeEncoder,    // ISO8601 UTC 时间格式
//		EncodeDuration:   zapcore.SecondsDurationEncoder,
//		EncodeCaller:     zapcore.ShortCallerEncoder,
//		ConsoleSeparator: "|",
//	}
//	// 设置日志级别
//	atomicLevel = zap.NewAtomicLevel()
//	atomicLevel.SetLevel(level)
//
//	core := zapcore.NewCore(
//		zapcore.NewConsoleEncoder(encoderConfig),
//		// zapcore.NewJSONEncoder(encoderConfig),
//		write,
//		// zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&write)), // 打印到控制台和文件
//		atomicLevel,
//	)
//	// 开启开发模式，堆栈跟踪
//	caller := zap.AddCaller()
//	// 开启文件及行号
//	development := zap.Development()
//	// 构造日志
//	log = zap.New(core, caller, development).Sugar()
//}
