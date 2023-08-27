package initialize

import (
	"server/config"
	"server/exchange"
	"server/exchange/binanceFuture"
	"server/global"
)

func Exchange() {
	exchange.Exchanges = make(map[string]exchange.Exchange)
	for _, ex := range global.Config.Exchanges {
		if ex.Enabled {
			switch ex.Name {
			case "binance_future":
				exchange.CreateDataBase(ex.Name)
				exc := CreateBinanceFuture(ex)
				exchange.Exchanges[ex.Name] = &exc
			}
		}
	}
	for _, ex := range exchange.Exchanges {
		ex.Init()
		ex.InitExchangeInfo()
		if global.Config.Mahakala.UpdateKline {
			ex.UpdateKlinesWithProgress()
		}
	}
}

func CreateBinanceFuture(ex config.Exchange) binanceFuture.BinanceFuture {
	return binanceFuture.BinanceFuture{
		BaseExchange: exchange.BaseExchange{
			Name:    ex.Name,
			Alias:   "币安合约",
			BaseUrl: ex.BaseUrl,
			Enabled: ex.Enabled,
			DB:      exchange.SetDataBase(ex.Name),
		},
		ApiKey:    ex.MetaData["api-key"].(string),
		SecretKey: ex.MetaData["secret-key"].(string),
	}
}
