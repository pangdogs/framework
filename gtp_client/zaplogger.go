package gtp_client

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var (
	DefaultLogger      *zap.Logger
	DefaultAtomicLevel zap.AtomicLevel
)

func init() {
	DefaultLogger, DefaultAtomicLevel = newConsoleZapLogger(zap.DebugLevel, "\t", true)
}

func newConsoleZapLogger(level zapcore.Level, separator string, development bool) (*zap.Logger, zap.AtomicLevel) {
	// 日志级别设置器
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(level)

	// 日志编码配置器
	encoderConfig := zapcore.EncoderConfig{
		LevelKey:         "level",
		TimeKey:          "timestamp",
		CallerKey:        "caller",
		StacktraceKey:    "stacktrace",
		MessageKey:       "msg",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.LowercaseLevelEncoder,
		EncodeTime:       zapcore.ISO8601TimeEncoder,
		EncodeDuration:   zapcore.SecondsDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: separator,
	}

	// 创建日志
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		atomicLevel,
	)

	options := []zap.Option{zap.AddCaller(), zap.AddStacktrace(zap.DPanicLevel)}
	if development {
		options = append(options, zap.Development())
	}

	logger := zap.New(core, options...)

	return logger, atomicLevel
}
