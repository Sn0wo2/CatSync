//go:build catsync_all || feature_acme_dns01

package framework

import (
	"fmt"
	"strings"

	"github.com/go-acme/lego/v4/challenge"
)

type dnsProviderCtor func() (challenge.Provider, error)

var (
	dnsProviderByName = map[string]dnsProviderCtor{}

	buildTagHints = map[string]string{
		"cloudflare": "dns_cloudflare",
		"dnspod":     "dns_dnspod",
		"alidns":     "dns_alidns",
		"route53":    "dns_route53",
	}
)

func registerDNSProvider(name string, ctor dnsProviderCtor) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" || ctor == nil {
		return
	}

	dnsProviderByName[name] = ctor
}

func newDNSProviderByName(name string) (challenge.Provider, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	ctor, ok := dnsProviderByName[name]
	if !ok {
		if tag, ok := buildTagHints[name]; ok {
			return nil, fmt.Errorf("dns provider %q is not compiled in; rebuild with -tags %s", name, tag)
		}
		return nil, fmt.Errorf("dns provider %q is not compiled in", name)
	}

	return ctor()
}
