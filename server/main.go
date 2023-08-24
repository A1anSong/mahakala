package main

import (
	"server/initialize"
)

func main() {
	// 初始化配置文件
	initialize.Viper()
	// 初始化日志
	initialize.Zap()
	// 初始化数据库
	initialize.Gorm()
	// 初始化resty
	initialize.Resty()
	// 初始化Carbon
	initialize.Carbon()
	// 初始化交易所
	initialize.Exchange()
	// 初始化定时任务
	initialize.Cron()
	// 初始化gin
	initialize.Gin()
}
