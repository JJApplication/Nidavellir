package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger 创建新的日志记录器
func NewLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	logger, err := config.Build(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewCore(zapcore.NewJSONEncoder(config.EncoderConfig), os.Stdout, config.Level)
	}))
	if err != nil {
		panic(err)
	}

	return logger
}

// NewDevelopmentLogger 创建开发环境日志记录器
func NewDevelopmentLogger() *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	return logger
}
