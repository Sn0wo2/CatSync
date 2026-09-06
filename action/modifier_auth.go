package action

import (
	"fmt"
	"net/netip"
	"regexp"
	"slices"
	"strings"

	"github.com/Sn0wo2/CatSync/internal/util"
)

type AuthFallbackPolicy uint8

const (
	AuthFallbackPolicyNext AuthFallbackPolicy = iota
	AuthFallbackPolicyJump
)

type AuthModifier struct {
	headerRules  map[string][]*regexp.Regexp
	queryRules   map[string]*regexp.Regexp
	ipWhiteList  *IPWhiteList
	fallback     AuthFallbackPolicy
	fallbackJump int
}

func (m *AuthModifier) Before(p *ProcessData) (*ProcessData, ExecutionResult) {
	logger := p.CStx.Logger

	if m.ipWhiteList != nil {
		ips := p.FCtx.IPs()
		if len(ips) == 0 {
			ips = []string{p.FCtx.IP()}
		}

		if !m.ipWhiteList.IsAllow(ips) {
			logger.Info("Auth >> IP not allowed",
				"actual", ips,
				"ctx", util.LazyFiberContext(p.FCtx),
			)

			return p, m.HandleFallback()
		}
	}

	reqHeaders := p.FCtx.GetReqHeaders()

	for k, rules := range m.headerRules {
		var values []string

		for hk, hv := range reqHeaders {
			if strings.EqualFold(hk, k) {
				values = hv

				break
			}
		}

		if len(values) == 0 {
			logger.Info("Auth >> Header missing", "header", k, "ctx", util.LazyFiberContext(p.FCtx))

			return p, m.HandleFallback()
		}

		if !slices.ContainsFunc(values, func(v string) bool {
			for _, re := range rules {
				if re.MatchString(v) {
					return true
				}
			}

			return false
		}) {
			patterns := make([]string, 0, len(rules))
			for _, re := range rules {
				patterns = append(patterns, re.String())
			}

			logger.Info("Auth >> Header value not matched",
				"header", k,
				"patterns", patterns,
				"actual", values,
				"ctx", util.LazyFiberContext(p.FCtx),
			)

			return p, m.HandleFallback()
		}
	}

	for k, rule := range m.queryRules {
		actual := p.FCtx.Query(k)

		if !rule.MatchString(actual) {
			logger.Info("Auth >> Query value not matched",
				"key", k,
				"pattern", rule.String(),
				"actual", actual,
				"ctx", util.LazyFiberContext(p.FCtx),
			)

			return p, m.HandleFallback()
		}
	}

	return p, ExecutionResult{}
}

func (m *AuthModifier) After(p *ProcessData) (*ProcessData, ExecutionResult) {
	return p, ExecutionResult{}
}

func (m *AuthModifier) HandleFallback() ExecutionResult {
	if m.fallback == AuthFallbackPolicyJump {
		return ExecutionResult{Status: ExecutionFallbackJump, JumpTo: m.fallbackJump}
	}

	return ExecutionResult{Status: ExecutionFallbackNext}
}

type IPWhiteList struct {
	addrs    []netip.Addr
	prefixes []netip.Prefix
}

func ParseIPWhiteList(entries []string) (*IPWhiteList, error) {
	wl := &IPWhiteList{}

	for i, raw := range entries {
		s := strings.TrimSpace(raw)
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}

		if strings.Contains(s, "/") {
			prefix, err := netip.ParsePrefix(s)
			if err != nil {
				return nil, fmt.Errorf("entry %d %q: %w", i, raw, err)
			}

			wl.prefixes = append(wl.prefixes, prefix)

			continue
		}

		addr, err := netip.ParseAddr(s)
		if err != nil {
			return nil, fmt.Errorf("entry %d %q: %w", i, raw, err)
		}

		wl.addrs = append(wl.addrs, addr)
	}

	return wl, nil
}

func (wl *IPWhiteList) IsAllow(ips []string) bool {
	for _, raw := range ips {
		ipStr := strings.TrimSpace(raw)
		if ipStr == "" {
			continue
		}

		addr, err := netip.ParseAddr(ipStr)
		if err != nil {
			continue
		}

		if slices.Contains(wl.addrs, addr) {
			return true
		}

		for _, prefix := range wl.prefixes {
			if prefix.Contains(addr) {
				return true
			}
		}
	}

	return false
}
