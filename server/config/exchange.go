package config

type Exchange struct {
	Name     string         `mapstructure:"name" json:"name" yaml:"name"`
	BaseUrl  string         `mapstructure:"base-url" json:"base-url" yaml:"base-url"`
	Enabled  bool           `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	MetaData map[string]any `mapstructure:"meta-data" json:"meta-data" yaml:"meta-data"`
}
