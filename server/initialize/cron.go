package initialize

import (
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"server/exchange"
	"server/global"
	"time"
)

func Cron() {
	utc, err := time.LoadLocation("UTC")
	if err != nil {
		global.Zap.Error("加载时区失败", zap.Error(err))
		panic(err)
	}
	c := cron.New(cron.WithLocation(utc))
	global.Cron = c
	setCron()
	global.Cron.Start()
}

func setCron() {
	_, err := global.Cron.AddFunc("*/30 * * * *", exchange.UpdateKlines)
	if err != nil {
		global.Zap.Error("定时任务运行错误", zap.Error(err))
		panic(err)
	}
}
