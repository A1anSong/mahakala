package config

type DB struct {
	Host     string `json:"host" form:"host" yaml:"host"`
	User     string `json:"user" form:"user" yaml:"user"`
	Password string `json:"password" form:"password" yaml:"password"`
	Dbname   string `json:"dbname" form:"dbname" yaml:"dbname"`
	Port     string `json:"port" form:"port" yaml:"port"`
	Config   string `json:"config" form:"config" yaml:"config"`
}
