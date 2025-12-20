package config

import (
	"errors"
	"net/http"
)

var (
	Path              string
	ErrConfigNotFound = errors.New("no config file found in default search paths")
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
	Status         uint16      `json:"status"         optional:"true" yaml:"status"`
	ResponseHeader http.Header `json:"responseHeader" optional:"true" yaml:"responseHeader"`
	Auth           ActionAuth  `json:"auth"           optional:"true" yaml:"auth"`

	Type ActionType `json:"type" yaml:"type"`

	// --- Action Data ---
	ActionFile   *ActionFileData   `json:"file"   optional:"true" yaml:"file"`
	ActionString *ActionStringData `json:"string" optional:"true" yaml:"string"`
}

type ActionAuth struct {
	Header map[string][]string `json:"header" optional:"true" yaml:"header"`
	Query  map[string]string   `json:"query"  optional:"true" yaml:"query"`
}

type ActionType string

const (
	ActionFile   ActionType = "file"
	ActionString ActionType = "string"
)

type ActionData interface {
	action()
}

type ActionFileData struct {
	ActionVersionModifier `json:"actionVersionModifier" optional:"true" yaml:"actionVersionModifier"`
	Path                  string `json:"path"                  yaml:"path"`
	DontDetectContentType bool   `json:"dontDetectContentType" optional:"true" yaml:"dontDetectContentType"`
}

func (a *ActionFileData) action() {}

type ActionStringData struct {
	ActionVersionModifier `json:"actionVersionModifier" optional:"true" yaml:"actionVersionModifier"`
	Content               string `json:"content"               yaml:"content"`
}

func (a *ActionStringData) action() {}

type ActionVersionModifier struct {
	Placeholder string `json:"placeholder" yaml:"placeholder"`
}
