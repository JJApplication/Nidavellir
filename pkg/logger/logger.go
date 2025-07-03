package logger

import (
	"go.uber.org/zap"
)

// NewLogger 创建新的日志记录器
func NewLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	logger, err := config.Build()
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
