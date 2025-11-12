package config

import (
	"errors"
	"net/http"
)

var (
	Path              string
	ErrConfigNotFound = errors.New("no config file found in default search paths")
)

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
			Operation:  OperationString,
			ActionData: "Hello, CatSync!",
		},
		{
			Route:     "/version",
			Operation: OperationVersion,
			ActionData: struct {
				Msg  string `json:"msg"  yaml:"msg"`
				Data any    `json:"data" yaml:"data"`
			}{
				Msg: "Hello, CatSync!",
				Data: map[string]string{
					"version": "{{version}}",
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
	Route          string          `json:"route"          yaml:"route"`
	Action         ActionType      `json:"action"         optional:"true"   yaml:"action"`
	Operation      ActionOperation `json:"operation"      optional:"true"   yaml:"operation"`
	ActionData     any             `json:"actionData"     yaml:"actionData"`
	ResponseHeader http.Header     `json:"responseHeader" optional:"true"   yaml:"responseHeader"`
	Auth           ActionAuth      `json:"auth"           optional:"true"   yaml:"auth"`
}

// Deprecated: Use config.ActionOperation instead.
type ActionType int

const (
	File = iota
	String
	TempRedirect
	Redirect
	JSON
)

func (t *ActionType) ToOperation() ActionOperation {
	switch *t {
	case String:
		return OperationString
	case TempRedirect:
		return OperationTempRedirect
	case Redirect:
		return OperationRedirect
	case JSON:
		return OperationJSON
	default: // 0(default): File
		return OperationFile
	}
}

type ActionOperation string

const (
	// System
	OperationVersion ActionOperation = "system-version"
	OperationReload  ActionOperation = "system-reload"

	// Actions
	OperationFile         ActionOperation = "file"
	OperationString       ActionOperation = "string"
	OperationTempRedirect ActionOperation = "temp_redirect"
	OperationRedirect     ActionOperation = "redirect"
	OperationJSON         ActionOperation = "json"
)

type ActionVersion struct {
	Msg  string `json:"msg"  yaml:"msg"`
	Data any    `json:"data" yaml:"data"`
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
