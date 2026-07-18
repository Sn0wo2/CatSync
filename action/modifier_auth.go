package action

import (
	"context"
	"errors"
	"net/netip"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/internal/util"
	"go.uber.org/zap"
)

type AuthModifier struct {
	auth config.ActionModifierAuth

	reOnce   sync.Once
	headerRE map[string][]*regexp.Regexp
	queryRE  map[string]*regexp.Regexp
	reErr    error

	ipOnce sync.Once
	ipWL   *ipAllowlist
	ipErr  error
}

type ipAllowlist struct {
	addrs    []netip.Addr
	prefixes []netip.Prefix
}

func NewAuthModifier(auth config.ActionModifierAuth) *AuthModifier {
	return &AuthModifier{auth: auth}
}

func (m *AuthModifier) Before(p *ProcessData) (*ProcessData, ExecutionResult) {
	logger := p.Ctx.Logger

	m.reOnce.Do(m.initRegex)

	if m.reErr != nil {
		logger.Warn("Auth >> Failed to init auth regex",
			zap.Error(m.reErr),
			util.LazyFiberContext(p.C),
		)

		return p, m.handleFallback()
	}

	m.ipOnce.Do(func() { m.initIPAllowlist(logger) })

	if m.ipErr != nil {
		logger.Warn("Auth >> Failed to init ip allowlist",
			zap.Error(m.ipErr),
			util.LazyFiberContext(p.C),
		)

		return p, m.handleFallback()
	}

	if m.ipWL != nil {
		ips := p.C.IPs()
		if len(ips) == 0 {
			ips = []string{p.C.IP()}
		}

		if !m.ipWL.allowAny(ips) {
			logger.Info("Auth >> IP not allowed",
				zap.Any("allowed", m.auth.IPAllow),
				zap.Strings("actual", ips),
				util.LazyFiberContext(p.C),
			)

			return p, m.handleFallback()
		}
	}

	reqHeaders := p.C.GetReqHeaders()

	for k, res := range m.headerRE {
		var values []string

		for hk, hv := range reqHeaders {
			if strings.EqualFold(hk, k) {
				values = hv

				break
			}
		}

		if len(values) == 0 {
			logger.Info("Auth >> Header missing", zap.String("header", k), util.LazyFiberContext(p.C))

			return p, m.handleFallback()
		}

		if !m.matchHeaderValues(values, res) {
			logger.Info("Auth >> Header value not matched",
				zap.String("header", k),
				zap.Any("patterns", m.auth.Header[k]),
				zap.Any("actual", values),
				util.LazyFiberContext(p.C),
			)

			return p, m.handleFallback()
		}
	}

	for k, re := range m.queryRE {
		if !re.MatchString(p.C.Query(k)) {
			logger.Info("Auth >> Query value not matched",
				zap.String("key", k),
				zap.Any("pattern", m.auth.Query[k]),
				zap.String("actual", p.C.Query(k)),
				util.LazyFiberContext(p.C),
			)

			return p, m.handleFallback()
		}
	}

	return p, ExecutionResult{}
}

func (m *AuthModifier) After(p *ProcessData) (*ProcessData, ExecutionResult) {
	return p, ExecutionResult{}
}

func (m *AuthModifier) matchHeaderValues(values []string, res []*regexp.Regexp) bool {
	for _, v := range values {
		for _, re := range res {
			if re.MatchString(v) {
				return true
			}
		}
	}

	return false
}

func (m *AuthModifier) initRegex() {
	if len(m.auth.Header) > 0 {
		m.headerRE = make(map[string][]*regexp.Regexp, len(m.auth.Header))
		for k, patterns := range m.auth.Header {
			out := make([]*regexp.Regexp, 0, len(patterns))
			for _, p := range patterns {
				pat, ok := p.LiteralTrim()
				if !ok {
					m.reErr = errors.New("auth.header pattern must be literal string")

					return
				}

				re, err := util.GetCompiledRegexp(pat)
				if err != nil {
					m.reErr = err

					return
				}

				out = append(out, re)
			}

			m.headerRE[k] = out
		}
	}

	if len(m.auth.Query) > 0 {
		m.queryRE = make(map[string]*regexp.Regexp, len(m.auth.Query))
		for k, p := range m.auth.Query {
			pat, ok := p.LiteralTrim()
			if !ok {
				m.reErr = errors.New("auth.query pattern must be literal string")

				return
			}

			re, err := util.GetCompiledRegexp(pat)
			if err != nil {
				m.reErr = err

				return
			}

			m.queryRE[k] = re
		}
	}
}

func (m *AuthModifier) initIPAllowlist(logger *zap.Logger) {
	if len(m.auth.IPAllow) == 0 && m.auth.IPFile == nil {
		m.ipWL = nil
		m.ipErr = nil

		return
	}

	wl := &ipAllowlist{}

	add := func(raw string) error {
		s := strings.TrimSpace(raw)
		if s == "" || strings.HasPrefix(s, "#") {
			return nil
		}

		if strings.Contains(s, "/") {
			prefix, err := netip.ParsePrefix(s)
			if err != nil {
				return err
			}

			wl.prefixes = append(wl.prefixes, prefix)

			return nil
		}

		addr, err := netip.ParseAddr(s)
		if err != nil {
			return err
		}

		wl.addrs = append(wl.addrs, addr)

		return nil
	}

	for _, raw := range m.auth.IPAllow {
		if err := add(raw); err != nil {
			m.ipErr = err

			return
		}
	}

	if m.auth.IPFile != nil {
		lines, err := m.auth.IPFile.ReadLines(context.Background())
		if err != nil {
			m.ipErr = err

			return
		}

		for i, line := range lines {
			if err := add(line); err != nil {
				m.ipErr = err
				if logger != nil {
					logger.Warn("Auth >> Invalid ipAllowlistFile entry",
						zap.Int("line", i+1),
						zap.String("value", strings.TrimSpace(line)),
						zap.Error(err),
					)
				}

				return
			}
		}
	}

	m.ipWL = wl
}

func (wl *ipAllowlist) allowAny(ips []string) bool {
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

func (m *AuthModifier) handleFallback() ExecutionResult {
	if m.auth.Fallback == nil || m.auth.Fallback.Type == "" {
		return ExecutionResult{Status: ExecutionFallbackNext}
	}

	switch m.auth.Fallback.Type {
	case config.AuthFallbackJump:
		return ExecutionResult{Status: ExecutionFallbackJump, JumpTo: m.auth.Fallback.JumpTo}
	case config.AuthFallbackNext:
		return ExecutionResult{Status: ExecutionFallbackNext}
	default:
		return ExecutionResult{Status: ExecutionFallbackNext}
	}
}
