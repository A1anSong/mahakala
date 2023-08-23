package common

import (
	"github.com/golang-module/carbon/v2"
	"github.com/shopspring/decimal"
)

type Kline struct {
	Time   carbon.DateTimeMilli `json:"time" gorm:"primaryKey;type:timestamptz;not null"`
	Open   decimal.Decimal      `json:"open" gorm:"type:numeric;not null"`
	High   decimal.Decimal      `json:"high" gorm:"type:numeric;not null"`
	Low    decimal.Decimal      `json:"low" gorm:"type:numeric;not null"`
	Close  decimal.Decimal      `json:"close" gorm:"type:numeric;not null"`
	Volume decimal.Decimal      `json:"volume" gorm:"type:numeric;not null"`
}
