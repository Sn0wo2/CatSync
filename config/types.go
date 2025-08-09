package config

import (
	"github.com/Sn0wo2/CatSync/action"
)

var Instance *Config

var DefaultConfig = &Config{
	Log: Log{
		Level: "debug",
		Dir:   "./logs",
	},
	Server: Server{
		Address: ":3000",
		Header:  "CatSync",
	},
	Actions: []Action{
		{
			Route:      "/",
			Action:     action.String,
			ActionData: "Hello CatSync",
		},
	},
}

type Config struct {
	IsDefault  bool   `json:"-" yaml:"-"`
	ConfigPath string `json:"-" yaml:"-"`

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
	UA         string      `json:"ua"         optional:"true"   yaml:"ua"`
}

type Loader interface {
	Load(cfg *Config, fileName string) error
	Save(cfg *Config, fileName string) error
	// GetAllowFileExtensions lowercase
	GetAllowFileExtensions() []string
}
