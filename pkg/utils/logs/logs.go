package logs

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	Development = 0
	Production
)

const loggerType = Development
var logger *zap.Logger

func createDevelopment() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
    return zap.Must(config.Build())
}

func init() {
	if loggerType == Development {
		logger = createDevelopment()
	}
}

func Debug(msg string, fields ...zapcore.Field) {
	logger.Debug(msg, fields...)
}
func Info(msg string, fields ...zapcore.Field) {
	logger.Info(msg, fields...)
}
func Warn(msg string, fields ...zapcore.Field) {
	logger.Warn(msg, fields...)
}
func Error(msg string, fields ...zapcore.Field) {
	logger.Error(msg, fields...)
}