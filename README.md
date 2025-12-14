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
	Header  string     `json:"header"  optional:"true" yaml:"header"`
	TLS     *ServerTLS `json:"tls"     optional:"true" yaml:"tls"`
}

type ServerTLS struct {
	Cert string `json:"cert" yaml:"cert"`
	Key  string `json:"key"  yaml:"key"`
}

type Action struct {
	Route          string       `json:"route"          yaml:"route"`
	ResponseHeader *http.Header `json:"responseHeader" optional:"true" yaml:"responseHeader"`
	Auth           *ActionAuth  `json:"auth"           optional:"true" yaml:"auth"`

	Type ActionType `json:"type" yaml:"type"`

	// --- Action Data ---
	ActionFile   *ActionFileData   `json:"file"   optional:"true" yaml:"file"`
	ActionString *ActionStringData `json:"string" optional:"true" yaml:"string"`
}

type ActionAuth struct {
	UA    string           `json:"ua"    optional:"true" yaml:"ua"`
	Query *ActionAuthQuery `json:"query" optional:"true" yaml:"query"`
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

type ActionData interface {
	data()
}

type ActionFileData struct {
	ActionVersionModifier `json:"actionVersionModifier" optional:"true" yaml:"actionVersionModifier"`
	Path                  string `json:"path"                  yaml:"path"`
}

func (a *ActionFileData) data() {}

type ActionStringData struct {
	ActionVersionModifier `json:"actionVersionModifier" optional:"true" yaml:"actionVersionModifier"`
	Content               string `json:"content"               yaml:"content"`
}

func (a *ActionStringData) data() {}

type ActionVersionModifier struct {
	Placeholder string `json:"placeholder" yaml:"placeholder"`
}
```

## Docker

### Run Docker

```bash

# Recommend
make run_docker

# OR: Replace 'latest' with 'local' to use your self-built images

docker run -d -p 3000:3000 -v ./data:/app/data:ro --name catsync catsync:latest
```

### Build Docker image

```bash

# Recommend
make test_goreleaser

# OR

go build -o CatSync ./cmd
docker build -t catsync:local .
```

## License

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FSn0wo2%2FCatSync.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FSn0wo2%2FCatSync?ref=badge_large)
