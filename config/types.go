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
	Log       Log              `json:"log"                 yaml:"log"`
	Server    Server           `json:"server"              yaml:"server"`
	Modifiers []GlobalModifier `json:"modifiers,omitempty" optional:"true" yaml:"modifiers,omitempty"`
	Actions   []Action         `json:"actions"             yaml:"actions"`
}

type Log struct {
	Dir        string `json:"dir"        optional:"true"   yaml:"dir"`
	Level      string `json:"level"      optional:"true"   yaml:"level"`
	FileFormat string `json:"fileFormat" yaml:"fileFormat"`
}

type Server struct {
	Address string      `json:"address" yaml:"address"`
	TLS     ServerTLS   `json:"tls"     optional:"true" yaml:"tls"`
	ACME    *ServerACME `json:"acme,omitempty" optional:"true" yaml:"acme,omitempty"`
}

// GlobalModifier
//
// Action 特有的Modifier会覆盖掉 GlobalModifier
type GlobalModifier struct {
	*ActionModifierResponseHeader `json:"actionModifierResponseHeader,omitempty" optional:"true" yaml:"actionModifierResponseHeader,omitempty"`
	*ActionModifierStatus         `json:"actionModifierStatus,omitempty"         optional:"true" yaml:"actionModifierStatus,omitempty"`
	*ActionModifierAuth           `json:"actionModifierAuth,omitempty"           optional:"true" yaml:"actionModifierAuth,omitempty"`
	*ActionModifierVersion        `json:"actionVersionModifier,omitempty"        optional:"true" yaml:"actionVersionModifier,omitempty"`
}

type ServerTLS struct {
	Cert string `json:"cert" yaml:"cert"`
	Key  string `json:"key"  yaml:"key"`

	// RedirectHTTP controls whether non-challenge HTTP requests are redirected to https.
	// When omitted, the default is true (only meaningful for ACME http-01).
	RedirectHTTP *bool `json:"redirectHttp,omitempty" optional:"true" yaml:"redirectHttp,omitempty"`
}

// ServerACME enables automatic TLS certificates via ACME (e.g. Let's Encrypt).
//
// Note: HTTP-01 challenge usually requires binding an HTTP address (default :80).
// If you run behind a reverse proxy, you can disable httpAddress and terminate
// the challenge upstream.
type ServerACME struct {
	Enable       bool     `json:"enable,omitempty"       optional:"true" yaml:"enable,omitempty"`
	Email        string   `json:"email,omitempty"        optional:"true" yaml:"email,omitempty"`
	CacheDir     string   `json:"cacheDir,omitempty"     optional:"true" yaml:"cacheDir,omitempty"`
	Hosts        []string `json:"hosts,omitempty"        optional:"true" yaml:"hosts,omitempty"`
	DirectoryURL string   `json:"directoryURL,omitempty" optional:"true" yaml:"directoryURL,omitempty"`
	HTTPAddress  string   `json:"httpAddress,omitempty"  optional:"true" yaml:"httpAddress,omitempty"`

	// challenge: http-01|dns-01 (default http-01)
	Challenge string         `json:"challenge,omitempty" optional:"true" yaml:"challenge,omitempty"`
	DNS       *ServerACMEDNS `json:"dns,omitempty" optional:"true" yaml:"dns,omitempty"`
}

// ServerACMEDNS configures DNS-01 challenge.
//
// This project uses lego under the hood for DNS-01.
// Provider options:
// - exec: run external commands to create/remove TXT records.
// - cloudflare/dnspod/alidns/route53: use lego built-in providers (may require build tags).
type ServerACMEDNS struct {
	Provider string `json:"provider,omitempty" optional:"true" yaml:"provider,omitempty"`

	// exec provider: command array (argv style). Supported placeholders:
	// {DOMAIN} {FQDN} {VALUE} {TOKEN} {KEYAUTH}
	PresentCmd []string `json:"presentCmd,omitempty" optional:"true" yaml:"presentCmd,omitempty"`
	CleanUpCmd []string `json:"cleanupCmd,omitempty" optional:"true" yaml:"cleanupCmd,omitempty"`

	PropagationTimeoutSeconds int `json:"propagationTimeoutSeconds,omitempty" optional:"true" yaml:"propagationTimeoutSeconds,omitempty"`
	PollingIntervalSeconds    int `json:"pollingIntervalSeconds,omitempty"    optional:"true" yaml:"pollingIntervalSeconds,omitempty"`
}

type Action struct {
	Route string `json:"route" optional:"true" yaml:"route"`

	// SkipGlobalModifiers disables config-level modifiers for this action.
	//
	// Some endpoints need to be isolated from defaults like global auth/headers,
	// while still benefiting from the per-action modifier pipeline.
	SkipGlobalModifiers bool `json:"skipGlobalModifiers,omitempty" optional:"true" yaml:"skipGlobalModifiers,omitempty"`

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

	Path               string `json:"path"               yaml:"path"`
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
	Header   map[string][]string         `json:"header,omitempty"   optional:"true" yaml:"header,omitempty"`
	Query    map[string]string           `json:"query,omitempty"    optional:"true" yaml:"query,omitempty"`
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
