package zap

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

// NewZapConsoleLogger 创建控制台样式日志记录器
func NewZapConsoleLogger(level zapcore.Level, separator, fileName string, maxSize int, stdout, development bool) (*zap.Logger, zap.AtomicLevel) {
	// 日志分割器与写入器
	rollingLogger := lumberjack.Logger{
		Filename: fileName, // 日志文件路径
		MaxSize:  maxSize,  // 每个日志文件大小
	}
	write := zapcore.AddSync(&rollingLogger)
	if stdout {
		write = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(write))
	}

	// 日志级别设置器
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(level)

	// 日志编码配置器
	encoderConfig := zapcore.EncoderConfig{
		LevelKey:         "level",
		TimeKey:          "time",
		CallerKey:        "caller",
		StacktraceKey:    "stacktrace",
		MessageKey:       "msg",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime:       zapcore.ISO8601TimeEncoder,    // ISO8601 UTC 时间格式
		EncodeDuration:   zapcore.NanosDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: separator,
	}

	// 创建日志
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		write,
		atomicLevel,
	)

	var options []zap.Option
	options = append(options, zap.AddCaller(), zap.AddStacktrace(zap.PanicLevel))

	if development {
		options = append(options, zap.Development())
	}

	logger := zap.New(core, options...)

	return logger, atomicLevel
}

// NewZapJsonLogger 创建Json样式日志记录器
func NewZapJsonLogger(level zapcore.Level, fileName string, maxSize int, stdout, development bool) (*zap.Logger, zap.AtomicLevel) {
	// 日志分割器与写入器
	rollingLogger := lumberjack.Logger{
		Filename: fileName, // 日志文件路径
		MaxSize:  maxSize,  // 每个日志文件大小
	}
	write := zapcore.AddSync(&rollingLogger)
	if stdout {
		write = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(write))
	}

	// 日志级别设置器
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(level)

	// 日志编码配置器
	encoderConfig := zapcore.EncoderConfig{
		LevelKey:       "level",
		TimeKey:        "time",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		MessageKey:     "msg",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.NanosDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建日志
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		write,
		atomicLevel,
	)

	var options []zap.Option
	options = append(options, zap.AddCaller(), zap.AddStacktrace(zap.PanicLevel))

	if development {
		options = append(options, zap.Development())
	}

	logger := zap.New(core, options...)

	return logger, atomicLevel
}
