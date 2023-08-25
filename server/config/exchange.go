package config

type Exchange struct {
	Name      string `mapstructure:"name" json:"name" yaml:"name"`
	BaseUrl   string `mapstructure:"base-url" json:"base-url" yaml:"base-url"`
	ApiKey    string `mapstructure:"api-key" json:"api-key" yaml:"api-key"`
	SecretKey string `mapstructure:"secret-key" json:"secret-key" yaml:"secret-key"`
	Enabled   bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
}
