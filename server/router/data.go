package router

import (
	"server/global"
	"server/service"
)

func SetDataRouter() {
	dataRouter := global.Gin.Group("data")
	{
		dataRouter.GET("", service.GetExchanges)
		dataRouter.GET("/:exchange", service.GetExchange)
		dataRouter.GET("/:exchange/:symbol", service.GetSymbolInfo)
		dataRouter.GET("/:exchange/:symbol/:interval", service.GetKlines)
	}
}
