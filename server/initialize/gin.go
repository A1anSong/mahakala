package initialize

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"server/global"
)

func Gin() {
	gin.SetMode(gin.ReleaseMode)
	global.Gin = gin.New()
	global.Gin.Use(gin.Logger(), gin.Recovery())
	global.Gin.GET("/ping", func(context *gin.Context) {
		context.JSON(200, gin.H{
			"message": "pong",
		})
	})
	err := global.Gin.Run(":8080")
	if err != nil {
		global.Zap.Error("gin运行错误", zap.Error(err))
		panic(err)
	}
}
