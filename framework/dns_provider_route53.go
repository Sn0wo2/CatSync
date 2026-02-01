//go:build dns_route53

package framework

import (
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/providers/dns/route53"
)

func init() {
	registerDNSProvider("route53", func() (challenge.Provider, error) {
		return route53.NewDNSProvider()
	})
}
