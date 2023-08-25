package router

import (
	"server/global"
	"server/service"
)

func SetPublicRouter() {
	publicRouter := global.Gin.Group("")
	{
		publicRouter.GET("/", service.Ping)
	}
}
