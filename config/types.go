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
}

type GlobalModifier struct {
	*ActionModifierResponseHeader `json:"actionModifierResponseHeader,omitempty" optional:"true" yaml:"actionModifierResponseHeader,omitempty"`
	*ActionModifierStatus         `json:"actionModifierStatus,omitempty"         optional:"true" yaml:"actionModifierStatus,omitempty"`
	*ActionModifierAuth           `json:"actionModifierAuth,omitempty"           optional:"true" yaml:"actionModifierAuth,omitempty"`
	*ActionModifierVersion        `json:"actionVersionModifier,omitempty"        optional:"true" yaml:"actionVersionModifier,omitempty"`
	*ActionModifierAge            `json:"actionModifierAge,omitempty"            optional:"true" yaml:"actionModifierAge,omitempty"`
}

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

	if gm.ActionModifierAge != nil {
		fn(gm.ActionModifierAge)
	}
}

func (gm *GlobalModifier) GetGlobalModifier() *GlobalModifier { return gm }

type ServerTLS struct {
	Cert *reader.String `json:"cert" yaml:"cert"`
	Key  *reader.String `json:"key"  yaml:"key"`
}

type Action struct {
	Route *reader.String `json:"route" optional:"true" yaml:"route"`

	Label string `json:"label,omitempty" optional:"true" yaml:"label,omitempty"`

	SkipGlobalModifiers bool `json:"skipGlobalModifiers,omitempty" optional:"true" yaml:"skipGlobalModifiers,omitempty"`

	Type *reader.String `json:"type" yaml:"type"`

	GlobalModifier `yaml:",inline"`

	ActionFile   *ActionFileData   `json:"file,omitempty"   optional:"true" yaml:"file,omitempty"`
	ActionString *ActionStringData `json:"string,omitempty" optional:"true" yaml:"string,omitempty"`
	ActionServer *ActionServerData `json:"server,omitempty" optional:"true" yaml:"server,omitempty"`
	ActionReload *ActionReloadData `json:"reload,omitempty" optional:"true" yaml:"reload,omitempty"`
}

func (a *Action) GetPayload() ActionData {
	switch a.TypeName() {
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

func (a *Action) TypeName() ActionType {
	if a.Type == nil {
		return ""
	}

	return ActionType(a.Type.Must())
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
	// TODO: Mode OR|AND
	Header      map[string][]*reader.String `json:"header,omitempty"          optional:"true" yaml:"header,omitempty"`
	Query       map[string]*reader.String   `json:"query,omitempty"           optional:"true" yaml:"query,omitempty"`
	IPWhiteList []string                    `json:"IPWhiteListlist,omitempty" optional:"true" yaml:"IPWhiteListlist,omitempty"`
	IPFile      *reader.String              `json:"ipAllowlistFile,omitempty" optional:"true" yaml:"ipAllowlistFile,omitempty"`
	Fallback    *ActionModifierAuthFallback `json:"fallback,omitempty"        optional:"true" yaml:"fallback,omitempty"`
}

type ActionModifierAuthFallback struct {
	Type ActionModifierAuthFallbackType `json:"type" yaml:"type"`

	JumpTo    int    `json:"jumpTo,omitempty"    optional:"true" yaml:"jumpTo,omitempty"`
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

type ActionModifierAge struct {
	// Recipients are age public keys (x25519 "age1..." from `mihomo age keygen`, or ssh public keys)
	Recipients []*reader.String `json:"recipients" yaml:"recipients"`
	// Armor outputs PEM-style age armor format, defaults to true
	Armor *bool `json:"armor,omitempty" optional:"true" yaml:"armor,omitempty"`
}

type Loader interface {
	GetTag() string
	Load(cfg *Config, fileName string) error
	Save(cfg *Config, fileName string) error
	GetAllowFileExtensions() []string
}
