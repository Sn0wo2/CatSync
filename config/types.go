package config

import (
	"errors"
	"net/http"
)

var (
	Path                 string
	ErrConfigNotFound    = errors.New("no config file found in default search paths")
	ErrInvalidActionData = errors.New("invalid action data")
)

type Config struct {
	Log     Log      `json:"log"     yaml:"log"`
	Server  Server   `json:"server"  yaml:"server"`
	Actions []Action `json:"actions" yaml:"actions"`
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
	Route          string      `json:"route"          yaml:"route"`
	ResponseHeader http.Header `json:"responseHeader" optional:"true" yaml:"responseHeader"`
	Auth           ActionAuth  `json:"auth"           optional:"true" yaml:"auth"`

	Type ActionType `json:"type" yaml:"type"`

	// --- Action Data ---
	ActionFile   *FileData   `json:"file,omitempty"   optional:"true" yaml:"file,omitempty"`
	ActionString *StringData `json:"string,omitempty" optional:"true" yaml:"string,omitempty"`
}

type ActionAuth struct {
	UA    string          `json:"ua"    optional:"true" yaml:"ua"`
	Query ActionAuthQuery `json:"query" optional:"true" yaml:"query"`
}

type ActionAuthQuery struct {
	Map            map[string]string `json:"map"            yaml:"map"`
	IgnoreCaseCase bool              `json:"ignoreCaseCase" optional:"true" yaml:"ignoreCaseCase"`
}

type ActionType string

const (
	ActionFile   ActionType = "file"
	ActionString ActionType = "string"
)

type ReloadData struct {
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
}

type FileData struct {
	Path string `json:"path" yaml:"path"`
}

type StringData struct {
	Content string `json:"content" yaml:"content"`
}
