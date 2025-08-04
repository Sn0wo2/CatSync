package config

import (
	"github.com/Sn0wo2/FileSync/internal/action"
)

var Instance *Config

type Config struct {
	Log     Log      `optional:"true" json:"log" yaml:"log"`
	Server  Server   `json:"server" yaml:"server"`
	Actions []Action `optional:"true" json:"actions" yaml:"actions"`
}

type Log struct {
	Level string `optional:"true" yaml:"level"`
	Dir   string `optional:"true" yaml:"dir"`
}

type Server struct {
	Address string `json:"address" yaml:"address"`
	Header  string `optional:"true" json:"header" yaml:"header"`
	TLS     TLS    `optional:"true" json:"tls" yaml:"tls"`
}

type TLS struct {
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

type Action struct {
	Route      string      `json:"route" yaml:"route"`
	Action     action.Type `json:"action" yaml:"action"`
	ActionData string      `json:"actionData" yaml:"actionData"`
}

type Loader interface {
	Load(fileName string) (*Config, error)
	GetAllowFileExtensions() []string
}
