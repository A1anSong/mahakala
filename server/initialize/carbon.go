package initialize

import (
	"github.com/golang-module/carbon/v2"
	"go.uber.org/zap"
	"server/global"
)

func Carbon() {
	lang := carbon.NewLanguage()
	lang.SetLocale("zh-CN")

	c := carbon.SetLanguage(lang)
	if c.Error != nil {
		global.Zap.Fatal("初始化 Carbon 时出错:", zap.Error(c.Error))
	}
	global.Carbon = c
}
