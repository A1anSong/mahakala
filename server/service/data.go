package service

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"server/exchange"
	"server/global"
	"server/utils"
)

func GetExchanges(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{
		"exchanges": exchange.GetExchanges(),
	})
}

func GetExchange(context *gin.Context) {
	reqExchange := context.Param("exchange")
	resExchange, exists := exchange.Exchanges[reqExchange]
	if !exists {
		context.JSON(http.StatusNotFound, gin.H{
			"message": "exchange not found",
		})
		return
	}
	context.JSON(200, resExchange)
}

func GetExchangeSymbols(context *gin.Context) {
	reqExchange := context.Param("exchange")
	resExchange, exists := exchange.Exchanges[reqExchange]
	if !exists {
		context.JSON(http.StatusNotFound, gin.H{
			"message": "exchange not found",
		})
		return
	}
	context.JSON(200, resExchange.GetSymbols())
}

func GetKlines(context *gin.Context) {
	reqExchange := context.Param("exchange")
	resExchange, exists := exchange.Exchanges[reqExchange]
	if !exists {
		context.JSON(http.StatusNotFound, gin.H{
			"message": "exchange not found",
		})
		return
	}
	symbol := context.Param("symbol")
	exists = resExchange.CheckSymbol(symbol)
	if !exists {
		context.JSON(http.StatusNotFound, gin.H{
			"message": "symbol not found",
		})
		return
	}
	interval := context.Param("interval")
	_, exists = utils.MapInterval[interval]
	if !exists {
		context.JSON(http.StatusNotFound, gin.H{
			"message": "not invalid interval",
		})
		return
	}
	klines, err := resExchange.GetKlines(symbol, interval)
	if err != nil {
		global.Zap.Error("获取K线错误", zap.Error(err))
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	context.JSON(200, klines)
}
