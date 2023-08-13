package initialize

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"server/global"
	"strings"
)

func Zap() {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)

	config.Encoding = "console"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(strings.ToUpper(level.String()))
	}

	logger, _ := config.Build()
	global.Zap = logger
}
