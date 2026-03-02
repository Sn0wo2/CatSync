# CatSync

> Sync the「cat」config backend server

---

[![Go Report Card](https://goreportcard.com/badge/github.com/Sn0wo2/CatSync)](https://goreportcard.com/report/github.com/Sn0wo2/CatSync)
[![GitHub release](https://img.shields.io/github/v/release/Sn0wo2/CatSync?color=blue)](https://github.com/Sn0wo2/CatSync/releases)
[![GitHub License](https://img.shields.io/github/license/Sn0wo2/CatSync)](LICENSE)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FSn0wo2%2FCatSync.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2FSn0wo2%2FCatSync?ref=badge_shield)

[![Go CI](https://github.com/Sn0wo2/CatSync/actions/workflows/go.yml/badge.svg)](https://github.com/Sn0wo2/CatSync/actions/workflows/go.yml)
[![Release](https://github.com/Sn0wo2/CatSync/actions/workflows/release.yml/badge.svg)](https://github.com/Sn0wo2/CatSync/actions/workflows/release.yml)
[![Dependabot Updates](https://github.com/Sn0wo2/CatSync/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/Sn0wo2/CatSync/actions/workflows/dependabot/dependabot-updates)
[![CodeQL Advanced](https://github.com/Sn0wo2/CatSync/actions/workflows/codeql.yml/badge.svg)](https://github.com/Sn0wo2/CatSync/actions/workflows/codeql.yml)

---

CatSync 正在努力向`V2`版本过渡, 尽可能保证功能的向后兼容性, 但是配置文件会有破坏性更新,
请根据下方配置类型和报错提示进行响应修改  
目前我们正在处于beta版本
如果您想体验请使用`docker run -d -p 3000:3000 -v ./data:/app/data:ro --name catsync ghcr.io/sn0wo2/catsync:beta-latest`
或者将`docker-compose.yml`中的`image`字段替换为`ghcr.io/sn0wo2/catsync:beta-latest`

或者从 https://github.com/Sn0wo2/CatSync/releases 下载最新的beta版本  
这是我们最新的更新计划: [TODO](./TODO.md)

> ### **⚠ BETA版本不是稳定版本, 不推荐在任何生产环境中使用。**

## Docs

- 快速开始：[`docs/quick-start.zh-CN.md`](docs/quick-start.zh-CN.md)
- 详细配置指南：[`docs/config-guide.zh-CN.md`](docs/config-guide.zh-CN.md)
- 文档索引：[`docs/README.md`](docs/README.md)

---

## Config & How 2 use

### All Config Types:

```go
// config/types.go
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
Address *reader.String `json:"address"        yaml:"address"`
TLS     ServerTLS      `json:"tls"            optional:"true" yaml:"tls"`
ACME    *ServerACME    `json:"acme,omitempty" optional:"true" yaml:"acme,omitempty"`
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
// - exec: run external commands to create/remove TXT records.
// - cloudflare/dnspod/alidns/route53: use lego built-in providers (may require build tags).
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
Route *reader.String `json:"route" optional:"true" yaml:"route"`

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
ActionServer *ActionServerData `json:"server,omitempty" optional:"true" yaml:"server,omitempty"`
ActionReload *ActionReloadData `json:"reload,omitempty" optional:"true" yaml:"reload,omitempty"`
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

// Modifier middleware

type ActionModifierResponseHeader struct {
	Header http.Header `json:"header" yaml:"header"`
}

type ActionModifierStatus struct {
	Status uint16 `json:"status" yaml:"status"`
}

type ActionModifierAuth struct {
Header   map[string][]*reader.String `json:"header,omitempty"          optional:"true" yaml:"header,omitempty"`
Query    map[string]*reader.String   `json:"query,omitempty"           optional:"true" yaml:"query,omitempty"`
IPAllow  []string                    `json:"ipAllowlist,omitempty"     optional:"true" yaml:"ipAllowlist,omitempty"`
IPFile   *reader.String              `json:"ipAllowlistFile,omitempty" optional:"true" yaml:"ipAllowlistFile,omitempty"`
Fallback *ActionModifierAuthFallback `json:"fallback,omitempty"        optional:"true" yaml:"fallback,omitempty"`
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
Placeholder *reader.String `json:"placeholder" yaml:"placeholder"`
}
```

## Docker

### Run Docker

```bash

# Recommend~
make run_docker

# Or Replace 'latest' with local image to use your self-built images

docker run -d -p 3000:3000 -v ./data:/app/data:ro --name catsync ghcr.io/sn0wo2/catsync:latest
```

### Build Docker image

```bash

# Recommend~
make test_goreleaser

# Or

go build -o CatSync ./cmd
docker build -t catsync:local .
```

## License

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FSn0wo2%2FCatSync.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FSn0wo2%2FCatSync?ref=badge_large)
