package config

type Server struct {
	DB DB `json:"db" form:"db" yaml:"db"`
}
