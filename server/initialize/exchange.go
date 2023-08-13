package initialize

import (
	"server/config"
	"server/exchange"
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
}

func CreateBinanceFuture(ex config.Exchange) exchange.BinanceFuture {
	return exchange.BinanceFuture{
		BaseExchange: exchange.BaseExchange{
			Name:      ex.Name,
			BaseUrl:   ex.BaseUrl,
			ApiKey:    "",
			SecretKey: "",
			DB:        exchange.SetDataBase(ex.Name),
		},
	}
}
