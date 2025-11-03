# CatSync

> Sync the「cat」config.

---

[![Go Report Card](https://goreportcard.com/badge/github.com/Sn0wo2/CatSync)](https://goreportcard.com/report/github.com/Sn0wo2/CatSync)
[![GitHub release](https://img.shields.io/github/v/release/Sn0wo2/CatSync?color=blue)](https://github.com/Sn0wo2/CatSync/releases)
[![GitHub License](https://img.shields.io/github/license/Sn0wo2/CatSync)](LICENSE)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FSn0wo2%2FCatSync.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2FSn0wo2%2FCatSync?ref=badge_shield)

[![Go CI](https://github.com/Sn0wo2/CatSync/actions/workflows/ci.yml/badge.svg)](https://github.com/Sn0wo2/CatSync/actions/workflows/ci.yml)
[![Release](https://github.com/Sn0wo2/CatSync/actions/workflows/release.yml/badge.svg)](https://github.com/Sn0wo2/CatSync/actions/workflows/release.yml)
[![Dependabot Updates](https://github.com/Sn0wo2/CatSync/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/Sn0wo2/CatSync/actions/workflows/dependabot/dependabot-updates)
[![CodeQL Advanced](https://github.com/Sn0wo2/CatSync/actions/workflows/codeql.yml/badge.svg)](https://github.com/Sn0wo2/CatSync/actions/workflows/codeql.yml)

---

## Config & How 2 use

### Default Config:

```yaml
log:
  level: debug
  dir: ./logs
server:
  address: :3000
  header: CatSync
  tls:
    cert: ""
    key: ""
actions:
  - route: /
    action: 1
    actionData: Hello CatSync
    responseHeader: { }
    auth:
      ua: ""
      query:
        map: { }
        ignoreCaseCase: false

```

---

### All Config Types:

```go
// config/types.go
type Config struct {
	Log     Log      `json:"log"     optional:"true" yaml:"log"`
	Server  Server   `json:"server"  yaml:"server"`
	Actions []Action `json:"actions" optional:"true" yaml:"actions"`

	// --- INTERNAL ---
	
	ConfigPath string `json:"-" optional:"true" yaml:"-"`
}

type Log struct {
	Level string `json:"level" optional:"true" yaml:"level"`
	Dir   string `json:"dir"   optional:"true" yaml:"dir"`
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
	Action         action.Type `json:"action"         yaml:"action"`
	ActionData     string      `json:"actionData"     yaml:"actionData"`
	ResponseHeader http.Header `json:"responseHeader" optional:"true"   yaml:"responseHeader"`
	Auth           ActionAuth  `json:"auth"           optional:"true"   yaml:"auth"`
}

type ActionAuth struct {
	UA    string          `json:"ua"    optional:"true" yaml:"ua"`
	Query ActionAuthQuery `json:"query" optional:"true" yaml:"query"`
}

type ActionAuthQuery struct {
	Map            map[string]string `json:"map"            yaml:"map"`
	IgnoreCaseCase bool              `json:"ignoreCaseCase" optional:"true" yaml:"ignoreCaseCase"`
}
```

## Docker

### Run Docker

```bash
docker-compose -f docker/docker-compose.yml up -d

# OR(Replace 'latest' to 'local'  use your self build images)

docker run -d -p 3000:3000 -v ./data:/app/data:ro --name catsync catsync:latest
```

### Build Docker image

```bash
goreleaser release --snapshot --clean

# OR

go build -o CatSync ./cmd
docker build -t catsync:local .
```
## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FSn0wo2%2FCatSync.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FSn0wo2%2FCatSync?ref=badge_large)
