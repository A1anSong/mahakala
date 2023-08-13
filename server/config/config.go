package config

type Server struct {
	DB        DB         `mapstructure:"db" json:"db" yaml:"db"`
	Exchanges []Exchange `mapstructure:"exchanges" json:"exchanges" yaml:"exchanges"`
}
