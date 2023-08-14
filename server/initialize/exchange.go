package initialize

import (
	"server/config"
	"server/exchange"
	"server/exchange/binanceFuture"
	"server/global"
)

func Exchange() {
	for _, ex := range global.Config.Exchanges {
		switch ex.Name {
		case "binance_future":
			exchange.CreateDataBase(ex.Name)
			exc := CreateBinanceFuture(ex)
			exchange.Exchanges = append(exchange.Exchanges, &exc)
		}
	}
	for _, ex := range exchange.Exchanges {
		ex.Init()
		ex.UpdateExchangeInfo()
	}
}

func CreateBinanceFuture(ex config.Exchange) binanceFuture.BinanceFuture {
	return binanceFuture.BinanceFuture{
		BaseExchange: exchange.BaseExchange{
			Name:      ex.Name,
			BaseUrl:   ex.BaseUrl,
			ApiKey:    ex.ApiKey,
			SecretKey: ex.SecretKey,
			DB:        exchange.SetDataBase(ex.Name),
		},
	}
}
