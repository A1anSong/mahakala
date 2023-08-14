package config

type Server struct {
	Mahakala  Mahakala   `mapstructure:"mahakala" json:"mahakala" yaml:"mahakala"`
	DB        DB         `mapstructure:"db" json:"db" yaml:"db"`
	Exchanges []Exchange `mapstructure:"exchanges" json:"exchanges" yaml:"exchanges"`
}
