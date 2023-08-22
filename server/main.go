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
	// 初始化交易所
	initialize.Exchange()
}
