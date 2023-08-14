package config

type Mahakala struct {
	KlineInterval    string `mapstructure:"kline-interval" json:"kline-interval" yaml:"kline-interval"`
	MaxUpdateRoutine int    `mapstructure:"max-update-routine" json:"max-update-routine" yaml:"max-update-routine"`
}
