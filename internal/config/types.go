package config

import (
	"github.com/Sn0wo2/FileSync/internal/action"
)

var Instance *Config

type Config struct {
	Log     Log      `json:"log"     optional:"true" yaml:"log"`
	Server  Server   `json:"server"  yaml:"server"`
	Actions []Action `json:"actions" optional:"true" yaml:"actions"`
}

type Log struct {
	Level string `json:"level" optional:"true" yaml:"level"`
	Dir   string `json:"dir"   optional:"true" yaml:"dir"`
}

type Server struct {
	Address string `json:"address" yaml:"address"`
	Header  string `json:"header"  optional:"true" yaml:"header"`
	TLS     TLS    `json:"tls"     optional:"true" yaml:"tls"`
}

type TLS struct {
	Cert string `json:"cert" yaml:"cert"`
	Key  string `json:"key"  yaml:"key"`
}
type Action struct {
	Route      string      `json:"route"      yaml:"route"`
	Action     action.Type `json:"action"     yaml:"action"`
	ActionData string      `json:"actionData" yaml:"actionData"`
}

type Loader interface {
	Load(fileName string) (*Config, error)
	GetAllowFileExtensions() []string
}
