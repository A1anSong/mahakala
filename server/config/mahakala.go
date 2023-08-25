package config

type Mahakala struct {
	Port             int    `mapstructure:"port" json:"port" yaml:"port"`
	UpdateKline      bool   `mapstructure:"update-kline" json:"update-kline" yaml:"update-kline"`
	KlineInterval    string `mapstructure:"kline-interval" json:"kline-interval" yaml:"kline-interval"`
	MaxUpdateRoutine int    `mapstructure:"max-update-routine" json:"max-update-routine" yaml:"max-update-routine"`
	AnalyzeAmount    int    `mapstructure:"analyze-amount" json:"analyze-amount" yaml:"analyze-amount"`
}
