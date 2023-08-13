package config

type DB struct {
	Host     string `mapstructure:"host" json:"host" yaml:"host"`
	User     string `mapstructure:"user" json:"user" yaml:"user"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
	Dbname   string `mapstructure:"dbname" json:"dbname" yaml:"dbname"`
	Port     string `mapstructure:"port" json:"port" yaml:"port"`
	Config   string `mapstructure:"config" json:"config" yaml:"config"`
}
