//go:build dns_dnspod

package framework

import (
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/providers/dns/dnspod"
)

func init() {
	registerDNSProvider("dnspod", func() (challenge.Provider, error) {
		return dnspod.NewDNSProvider()
	})
}
