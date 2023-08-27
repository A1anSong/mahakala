package response

import (
	"github.com/golang-module/carbon/v2"
	"github.com/shopspring/decimal"
)

type Kline struct {
	Period carbon.Timestamp `json:"time"`
	Open   decimal.Decimal  `json:"open"`
	High   decimal.Decimal  `json:"high"`
	Low    decimal.Decimal  `json:"low"`
	Close  decimal.Decimal  `json:"close"`
	Volume decimal.Decimal  `json:"volume"`
}
