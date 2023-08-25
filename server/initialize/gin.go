package initialize

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"server/global"
	"server/router"
)

func Gin() {
	gin.SetMode(gin.ReleaseMode)
	global.Gin = gin.New()
	global.Gin.Use(gin.Logger(), gin.Recovery())
	router.SetRouter()
	err := global.Gin.Run(fmt.Sprintf(":%d", global.Config.Mahakala.Port))
	if err != nil {
		global.Zap.Error("gin运行错误", zap.Error(err))
		panic(err)
	}
}
