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
	Log     Log      `json:"log"                 yaml:"log"`
	Server  Server   `json:"server"              yaml:"server"`
	Modifiers []GlobalModifier `json:"modifiers,omitempty" optional:"true" yaml:"modifiers,omitempty"`
	Actions []Action `json:"actions"             yaml:"actions"`
}

type Log struct {
	Dir        string `json:"dir"        optional:"true"   yaml:"dir"`
	Level      string `json:"level"      optional:"true"   yaml:"level"`
	FileFormat string `json:"fileFormat" yaml:"fileFormat"`
}

type Server struct {
	Address string    `json:"address" yaml:"address"`
	TLS     ServerTLS `json:"tls"     optional:"true" yaml:"tls"`
}

// GlobalModifier
//
// Action 特有的Modifier会覆盖掉 GlobalModifier
type GlobalModifier struct {
	*ActionModifierResponseHeader `json:"actionModifierResponseHeader,omitempty" optional:"true" yaml:"actionModifierResponseHeader,omitempty"`
	*ActionModifierStatus  `json:"actionModifierStatus,omitempty"         optional:"true" yaml:"actionModifierStatus,omitempty"`
	*ActionModifierAuth    `json:"actionModifierAuth,omitempty"           optional:"true" yaml:"actionModifierAuth,omitempty"`
	*ActionModifierVersion `json:"actionVersionModifier,omitempty"        optional:"true" yaml:"actionVersionModifier,omitempty"`
}

type ServerTLS struct {
	Cert string `json:"cert" yaml:"cert"`
	Key  string `json:"key"  yaml:"key"`
}

type Action struct {
	Route string `json:"route" optional:"true" yaml:"route"`

	Type ActionType `json:"type" yaml:"type"`

	// --- Action Modifiers ---
	GlobalModifier `yaml:",inline"`

	// --- Action Datas ---
	ActionFile   *ActionFileData   `json:"file,omitempty"   optional:"true" yaml:"file,omitempty"`
	ActionString *ActionStringData `json:"string,omitempty" optional:"true" yaml:"string,omitempty"`
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
	GlobalModifier `yaml:",inline"`

	Path string `json:"path"               yaml:"path"`
	DontSetContentType bool   `json:"dontSetContentType" optional:"true" yaml:"dontSetContentType"`
}

func (a *ActionFileData) action() {}

type ActionStringData struct {
	GlobalModifier `yaml:",inline"`

	Content string `json:"content" yaml:"content"`
}

func (a *ActionStringData) action() {}

// Modifier middleware

type ActionModifierResponseHeader struct {
	Header http.Header `json:"header" yaml:"header"`
}

type ActionModifierStatus struct {
	Status uint16 `json:"status" yaml:"status"`
}

type ActionModifierAuth struct {
	Header map[string][]string `json:"header,omitempty"   optional:"true" yaml:"header,omitempty"`
	Query  map[string]string   `json:"query,omitempty"    optional:"true" yaml:"query,omitempty"`
	Fallback *ActionModifierAuthFallback `json:"fallback,omitempty" optional:"true" yaml:"fallback,omitempty"`
}

type ActionModifierAuthFallback struct {
	Type   ActionModifierAuthFallbackType `json:"type"             yaml:"type"`
	JumpTo int                            `json:"jumpTo,omitempty" optional:"true" yaml:"jumpTo,omitempty"`
}

type ActionModifierAuthFallbackType string

const (
	AuthFallbackJump ActionModifierAuthFallbackType = "jump"
	AuthFallbackNext ActionModifierAuthFallbackType = "next"
)

type ActionModifierVersion struct {
	Placeholder string `json:"placeholder" yaml:"placeholder"`
}
