package config

import (
	"net/http"

	"github.com/Sn0wo2/CatSync/config/reader"
)

type LoadResult struct {
	Config *Config
	Path   string
}

type Config struct {
	Log       Log              `json:"log"                 yaml:"log"`
	Server    Server           `json:"server"              yaml:"server"`
	Modifiers []GlobalModifier `json:"modifiers,omitempty" optional:"true" yaml:"modifiers,omitempty"`
	Actions   []Action         `json:"actions"             yaml:"actions"`
}

type Log struct {
	Dir        *reader.String `json:"dir"        optional:"true"   yaml:"dir"`
	Level      *reader.String `json:"level"      optional:"true"   yaml:"level"`
	FileFormat *reader.String `json:"fileFormat" yaml:"fileFormat"`
}

type Server struct {
	Address *reader.String `json:"address"           yaml:"address"`
	Prefork bool           `json:"prefork,omitempty" optional:"true" yaml:"prefork,omitempty"`
	TLS     ServerTLS      `json:"tls"               optional:"true" yaml:"tls"`
	ACME    *ServerACME    `json:"acme,omitempty"    optional:"true" yaml:"acme,omitempty"`
}

type GlobalModifier struct {
	*ActionModifierResponseHeader `json:"actionModifierResponseHeader,omitempty" optional:"true" yaml:"actionModifierResponseHeader,omitempty"`
	*ActionModifierStatus         `json:"actionModifierStatus,omitempty"         optional:"true" yaml:"actionModifierStatus,omitempty"`
	*ActionModifierAuth           `json:"actionModifierAuth,omitempty"           optional:"true" yaml:"actionModifierAuth,omitempty"`
	*ActionModifierVersion        `json:"actionVersionModifier,omitempty"        optional:"true" yaml:"actionVersionModifier,omitempty"`
}

// EachModifier calls fn for each non-nil modifier on this GlobalModifier.
// fn receives the concrete modifier type.
func (gm *GlobalModifier) EachModifier(fn func(any)) {
	if gm == nil {
		return
	}

	if gm.ActionModifierStatus != nil {
		fn(gm.ActionModifierStatus)
	}

	if gm.ActionModifierAuth != nil {
		fn(gm.ActionModifierAuth)
	}

	if gm.ActionModifierResponseHeader != nil {
		fn(gm.ActionModifierResponseHeader)
	}

	if gm.ActionModifierVersion != nil {
		fn(gm.ActionModifierVersion)
	}
}

func (gm *GlobalModifier) GetGlobalModifier() *GlobalModifier { return gm }

type ServerTLS struct {
	Cert *reader.String `json:"cert" yaml:"cert"`
	Key  *reader.String `json:"key"  yaml:"key"`

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
	Enable       bool           `json:"enable,omitempty"       optional:"true" yaml:"enable,omitempty"`
	Email        *reader.String `json:"email,omitempty"        optional:"true" yaml:"email,omitempty"`
	CacheDir     *reader.String `json:"cacheDir,omitempty"     optional:"true" yaml:"cacheDir,omitempty"`
	Hosts        []string       `json:"hosts,omitempty"        optional:"true" yaml:"hosts,omitempty"`
	DirectoryURL *reader.String `json:"directoryURL,omitempty" optional:"true" yaml:"directoryURL,omitempty"`

	// HTTP01 and DNS01 are mutually exclusive.
	// If both are nil, HTTP-01 is used with defaults.
	HTTP01 *ServerACMEHTTP01 `json:"http01,omitempty" optional:"true" yaml:"http01,omitempty"`
	DNS01  *ServerACMEDNS01  `json:"dns01,omitempty"  optional:"true" yaml:"dns01,omitempty"`
}

// ServerACMEHTTP01 configures HTTP-01 challenge.
type ServerACMEHTTP01 struct {
	HTTPAddress *reader.String `json:"httpAddress,omitempty" optional:"true" yaml:"httpAddress,omitempty"`
}

// ServerACMEDNS01 configures DNS-01 challenge.
//
// This project uses lego under the hood for DNS-01.
// Provider options:
//   - exec: run external commands to create/remove TXT records.
//   - cloudflare/dnspod/alidns/route53: use lego built-in providers (may require build tags).
type ServerACMEDNS01 struct {
	Provider *reader.String `json:"provider,omitempty" optional:"true" yaml:"provider,omitempty"`

	// exec provider: command array (argv style). Supported placeholders:
	// {DOMAIN} {FQDN} {VALUE} {TOKEN} {KEYAUTH}
	PresentCmd []string `json:"presentCmd,omitempty" optional:"true" yaml:"presentCmd,omitempty"`
	CleanUpCmd []string `json:"cleanupCmd,omitempty" optional:"true" yaml:"cleanupCmd,omitempty"`

	PropagationTimeoutSeconds int `json:"propagationTimeoutSeconds,omitempty" optional:"true" yaml:"propagationTimeoutSeconds,omitempty"`
	PollingIntervalSeconds    int `json:"pollingIntervalSeconds,omitempty"    optional:"true" yaml:"pollingIntervalSeconds,omitempty"`
}

type Action struct {
	// Label is an optional unique name for this action.
	// Used for label-based auth fallback jumps (jumpTo: "label").
	Label string `json:"label,omitempty" optional:"true" yaml:"label,omitempty"`

	Route *reader.String `json:"route" optional:"true" yaml:"route"`

	// SkipGlobalModifiers disables config-level modifiers for this action.
	//
	// Some endpoints need to be isolated from defaults like global auth/headers,
	// while still benefiting from the per-action modifier pipeline.
	SkipGlobalModifiers bool `json:"skipGlobalModifiers,omitempty" optional:"true" yaml:"skipGlobalModifiers,omitempty"`

	Type ActionType `json:"type" yaml:"type"`

	GlobalModifier `yaml:",inline"`

	ActionFile   *ActionFileData   `json:"file,omitempty"   optional:"true" yaml:"file,omitempty"`
	ActionString *ActionStringData `json:"string,omitempty" optional:"true" yaml:"string,omitempty"`
	ActionServer *ActionServerData `json:"server,omitempty" optional:"true" yaml:"server,omitempty"`
	ActionReload *ActionReloadData `json:"reload,omitempty" optional:"true" yaml:"reload,omitempty"`
}

func (a *Action) GetPayload() ActionData {
	switch a.Type {
	case ActionFile:
		if a.ActionFile != nil {
			return a.ActionFile
		}
	case ActionString:
		if a.ActionString != nil {
			return a.ActionString
		}
	case ActionServer:
		if a.ActionServer != nil {
			return a.ActionServer
		}
	case ActionReload:
		if a.ActionReload != nil {
			return a.ActionReload
		}
	}

	return nil
}

type ActionType string

const (
	ActionFile   ActionType = "file"
	ActionString ActionType = "string"
	ActionServer ActionType = "server"
	ActionReload ActionType = "reload"
)

type ActionData interface {
	action()
	GetGlobalModifier() *GlobalModifier
}

type ActionFileData struct {
	GlobalModifier `yaml:",inline"`

	Path               *reader.String `json:"path"               yaml:"path"`
	DontSetContentType bool           `json:"dontSetContentType" optional:"true" yaml:"dontSetContentType"`
}

func (a *ActionFileData) action() {}

type ActionServerData struct {
	GlobalModifier `yaml:",inline"`

	Directory    *reader.String   `json:"directory"              yaml:"directory"`
	IndexFiles   []*reader.String `json:"indexFiles,omitempty"   optional:"true"  yaml:"indexFiles,omitempty"`
	NotFoundHTML *reader.String   `json:"notFoundHTML,omitempty" optional:"true"  yaml:"notFoundHTML,omitempty"`
}

func (a *ActionServerData) action() {}

type ActionReloadData struct {
	GlobalModifier `yaml:",inline"`
}

func (a *ActionReloadData) action() {}

type ActionStringData struct {
	GlobalModifier `yaml:",inline"`

	Content *reader.String `json:"content" yaml:"content"`
}

func (a *ActionStringData) action() {}

type ActionModifierResponseHeader struct {
	Header   http.Header    `json:"header"             yaml:"header"`
	Upstream *reader.String `json:"upstream,omitempty" optional:"true" yaml:"upstream,omitempty"`
}

type ActionModifierStatus struct {
	Status   uint16         `json:"status"             yaml:"status"`
	Upstream *reader.String `json:"upstream,omitempty" optional:"true" yaml:"upstream,omitempty"`
}

type ActionModifierAuth struct {
	Header   map[string][]*reader.String `json:"header,omitempty"          optional:"true" yaml:"header,omitempty"`
	Query    map[string]*reader.String   `json:"query,omitempty"           optional:"true" yaml:"query,omitempty"`
	IPAllow  []string                    `json:"ipAllowlist,omitempty"     optional:"true" yaml:"ipAllowlist,omitempty"`
	IPFile   *reader.String              `json:"ipAllowlistFile,omitempty" optional:"true" yaml:"ipAllowlistFile,omitempty"`
	Fallback *ActionModifierAuthFallback `json:"fallback,omitempty"        optional:"true" yaml:"fallback,omitempty"`
}

type ActionModifierAuthFallback struct {
	Type ActionModifierAuthFallbackType `json:"type" yaml:"type"`

	// JumpTo is the action index to jump to (deprecated, use JumpLabel).
	JumpTo int `json:"jumpTo,omitempty" optional:"true" yaml:"jumpTo,omitempty"`

	// JumpLabel is the label of the target action (takes priority over JumpTo).
	JumpLabel string `json:"jumpLabel,omitempty" optional:"true" yaml:"jumpLabel,omitempty"`
}

type ActionModifierAuthFallbackType string

const (
	AuthFallbackJump ActionModifierAuthFallbackType = "jump"
	AuthFallbackNext ActionModifierAuthFallbackType = "next"
)

type ActionModifierVersion struct {
	Placeholder *reader.String `json:"placeholder" yaml:"placeholder"`
}
