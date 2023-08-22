package config

type DB struct {
	Host          string `mapstructure:"host" json:"host" yaml:"host"`
	Port          string `mapstructure:"port" json:"port" yaml:"port"`
	User          string `mapstructure:"user" json:"user" yaml:"user"`
	Password      string `mapstructure:"password" json:"password" yaml:"password"`
	InitialDBName string `mapstructure:"initial-db-name" json:"initial-db-name" yaml:"initial-db-name"`
	DefaultDBName string `mapstructure:"default-db-name" json:"default-db-name" yaml:"default-db-name"`
	Config        string `mapstructure:"config" json:"config" yaml:"config"`
}
