//go:build dns_cloudflare

package framework

import (
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
)

func init() {
	registerDNSProvider("cloudflare", func() (challenge.Provider, error) {
		return cloudflare.NewDNSProvider()
	})
}
