package router

import (
	"server/global"
	"server/service"
)

func SetDataRouter() {
	dataRouter := global.Gin.Group("data")
	{
		dataRouter.GET("/exchanges", service.GetExchanges)
		dataRouter.GET("/:exchange", service.GetExchange)
		dataRouter.GET("/:exchange/symbols", service.GetExchangeSymbols)
		dataRouter.GET("/:exchange/:symbol/klines/:interval", service.GetKlines)
	}
}
