package initialize

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"server/global"
	"strings"
)

func Zap() {
	config := zap.NewProductionConfig()

	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(strings.ToUpper(level.String()))
	}

	// 创建一个文件来存储日志
	if _, err := os.Stat("log"); os.IsNotExist(err) {
		errDir := os.MkdirAll("log", 0755)
		if errDir != nil {
			panic(errDir)
		}
	}

	// 创建一个文件来存储日志
	errorLogFile, errFile := os.OpenFile("log/error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if errFile != nil {
		panic(errFile)
	}

	// 创建两个core：一个用于写入错误日志到文件，另一个用于输出到控制台
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(config.EncoderConfig),
		zapcore.Lock(os.Stderr),
		zapcore.DebugLevel,
	)
	fileCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(config.EncoderConfig),
		zapcore.Lock(errorLogFile),
		zapcore.ErrorLevel,
	)

	// 使用tee core将两个core组合在一起
	core := zapcore.NewTee(consoleCore, fileCore)

	// 创建并设置logger
	logger := zap.New(core)
	global.Zap = logger
}
