package logger

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger 创建Zap日志实例
func NewLogger() (*zap.Logger, error) {
	var config zap.Config

	// 根据配置选择开发或生产环境的日志配置
	if viper.GetString("server.mode") == "debug" {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}

	// 创建日志实例
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}

// NewSugaredLogger 创建SugaredLogger实例
func NewSugaredLogger(logger *zap.Logger) *zap.SugaredLogger {
	return logger.Sugar()
}
