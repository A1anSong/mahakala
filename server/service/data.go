package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"server/exchange"
	"server/global"
	"server/service/response"
	"server/utils"
)

func GetExchanges(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"exchanges": exchange.GetExchanges(),
	})
}

func GetExchange(c *gin.Context) {
	reqExchange := c.Param("exchange")
	resExchange, exists := exchange.Exchanges[reqExchange]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "exchange not found",
		})
		return
	}
	exc := response.Exchange{
		Exchange: resExchange,
		Symbols:  resExchange.GetSymbols(),
	}
	c.JSON(200, exc)
}

func GetKlines(c *gin.Context) {
	reqExchange := c.Param("exchange")
	resExchange, exists := exchange.Exchanges[reqExchange]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "exchange not found",
		})
		return
	}
	symbol := c.Param("symbol")
	exists = resExchange.CheckSymbol(symbol)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "symbol not found",
		})
		return
	}
	interval := c.Param("interval")
	_, exists = utils.MapInterval[interval]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "not invalid interval",
		})
		return
	}
	klines, err := resExchange.GetKlines(symbol, interval)
	if err != nil {
		global.Zap.Error("获取K线错误", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	fmt.Println(klines[len(klines)-1].Period)
	c.JSON(200, klines)
}
