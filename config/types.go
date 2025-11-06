package config

import (
	"net/http"

	"github.com/Sn0wo2/CatSync/action"
	"github.com/Sn0wo2/CatSync/version"
	"github.com/gofiber/fiber/v2"
)

var Path string

var DefaultConfig = &Config{
	Log: Log{
		Level:      "debug",
		Dir:        "./logs",
		FileFormat: "2006-01-02.log",
	},
	Server: Server{
		Address: ":3000",
		Header:  "CatSync",
	},
	Actions: []Action{
		{
			Route:      "/",
			Action:     action.String,
			ActionData: "Hello from CatSync!",
		},
		{
			Route:  "/json",
			Action: action.JSON,
			ActionData: fiber.Map{
				"msg": "Hello from CatSync!",
				"data": fiber.Map{
					"version": version.GetFormatVersion(),
				},
			},
		},
	},
}

type Config struct {
	Log     Log      `json:"log"     yaml:"log"`
	Server  Server   `json:"server"  yaml:"server"`
	Actions []Action `json:"actions" optional:"true" yaml:"actions"`
}

type Log struct {
	Dir        string `json:"dir"        optional:"true"   yaml:"dir"`
	Level      string `json:"level"      optional:"true"   yaml:"level"`
	FileFormat string `json:"fileFormat" yaml:"fileFormat"`
}

type Server struct {
	Address string    `json:"address" yaml:"address"`
	Header  string    `json:"header"  optional:"true" yaml:"header"`
	TLS     ServerTLS `json:"tls"     optional:"true" yaml:"tls"`
}

type ServerTLS struct {
	Cert string `json:"cert" yaml:"cert"`
	Key  string `json:"key"  yaml:"key"`
}

type Action struct {
	Route          string           `json:"route"          yaml:"route"`
	Action         action.Type      `json:"action"         optional:"true"   yaml:"action"`
	Operation      action.Operation `json:"operation"      optional:"true"   yaml:"operation"`
	ActionData     action.Data      `json:"actionData"     yaml:"actionData"`
	ResponseHeader http.Header      `json:"responseHeader" optional:"true"   yaml:"responseHeader"`
	Auth           ActionAuth       `json:"auth"           optional:"true"   yaml:"auth"`
}

type ActionAuth struct {
	UA    string          `json:"ua"    optional:"true" yaml:"ua"`
	Query ActionAuthQuery `json:"query" optional:"true" yaml:"query"`
}

type ActionAuthQuery struct {
	Map            map[string]string `json:"map"            yaml:"map"`
	IgnoreCaseCase bool              `json:"ignoreCaseCase" optional:"true" yaml:"ignoreCaseCase"`
}

type Loader interface {
	GetTag() string
	Load(cfg *Config, fileName string) error
	Save(cfg *Config, fileName string) error
	// GetAllowFileExtensions lowercase
	GetAllowFileExtensions() []string
}
