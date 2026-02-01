//go:build dns_alidns

package framework

import (
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
)

func init() {
	registerDNSProvider("alidns", func() (challenge.Provider, error) {
		return alidns.NewDNSProvider()
	})
}
