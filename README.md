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
	Address string     `json:"address" yaml:"address"`
	TLS     *ServerTLS `json:"tls"     optional:"true" yaml:"tls"`
}

type ServerTLS struct {
	Cert string `json:"cert" yaml:"cert"`
	Key  string `json:"key"  yaml:"key"`
}

type Action struct {
	Route string `json:"route"          yaml:"route"`

	// --- Action Modifiers ---
	GlobalModifier

	Type ActionType `json:"type" yaml:"type"`

	// --- Action Data ---
	ActionFile   *ActionFileData   `json:"file"   optional:"true" yaml:"file"`
	ActionString *ActionStringData `json:"string" optional:"true" yaml:"string"`
}

type ActionType string

const (
	ActionFile   ActionType = "file"
	ActionString ActionType = "string"
)

type ActionData interface {
	data()
}

type ActionFileData struct {
	GlobalModifier
	Path                  string `json:"path"                  yaml:"path"`
}

func (a *ActionFileData) data() {}

type ActionStringData struct {
	GlobalModifier
	Content               string `json:"content"               yaml:"content"`
}

func (a *ActionStringData) data() {}

type GlobalModifier struct {
	*ActionModifierResponseHeader `json:"actionModifierResponseHeader" optional:"true" yaml:"actionModifierResponseHeader"`
	*ActionModifierStatus         `json:"actionModifierStatus" optional:"true" yaml:"actionModifierStatus"`
	*ActionModifierAuth           `json:"actionModifierAuth" optional:"true" yaml:"actionModifierAuth"`
	*ActionModifierVersion        `json:"actionVersionModifier" optional:"true" yaml:"actionVersionModifier"`
}

type ActionModifierResponseHeader struct {
	Header http.Header `json:"header" yaml:"header"`
}

type ActionModifierStatus struct {
	Status uint16 `json:"status" yaml:"status"`
}

type ActionModifierAuth struct {
	Header   map[string][]string `json:"header" optional:"true" yaml:"header"`
	Query    map[string]string   `json:"query"  optional:"true" yaml:"query"`
	Fallback *ActionModifierAuthFallback `json:"fallback" optional:"true" yaml:"fallback"`
}

type ActionModifierAuthFallback struct {
	Type   ActionModifierAuthFallbackType `json:"type" yaml:"type"`
	JumpTo uint                           `json:"jumpTo" optional:"true" yaml:"jumpTo"`
}

type ActionModifierAuthFallbackType string

const (
	AuthFallbackJump ActionModifierAuthFallbackType = "jump"
	AuthFallbackNext ActionModifierAuthFallbackType = "next"
)

type ActionModifierVersion struct {
	Placeholder string `json:"placeholder" yaml:"placeholder"`
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
