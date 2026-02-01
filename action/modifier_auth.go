package action

import (
	"context"
	"errors"
	"net"
	"regexp"
	"strings"
	"sync"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/config/reader"
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
	ips  []net.IP
	nets []*net.IPNet
}

func NewAuthModifier(auth config.ActionModifierAuth) *AuthModifier {
	return &AuthModifier{auth: auth}
}

func (m *AuthModifier) ProcessModifier(handler Handler) Handler {
	return WrapHandlerWithHooks(handler).Before(func(p *ProcessData) (*ProcessData, error) {
		logger := p.Ctx.GetLogger()

		m.reOnce.Do(m.initRegex)

		if m.reErr != nil {
			logger.Warn("Auth >> Failed to init auth regex",
				zap.Error(m.reErr),
				zap.String("ctx", util.FiberContextString(p.C)),
			)

			return m.handleFallback()
		}

		m.ipOnce.Do(func() { m.initIPAllowlist(logger) })

		if m.ipErr != nil {
			logger.Warn("Auth >> Failed to init ip allowlist",
				zap.Error(m.ipErr),
				zap.String("ctx", util.FiberContextString(p.C)),
			)

			return m.handleFallback()
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
					zap.String("ctx", util.FiberContextString(p.C)),
				)

				return m.handleFallback()
			}
		}

		// Header checks
		reqHeaders := p.C.GetReqHeaders()

		for k, res := range m.headerRE {
			var values []string

			for hk, hv := range reqHeaders {
				if strings.EqualFold(hk, k) {
					values = hv

					break
				}
			}

			// Missing header should be treated as auth mismatch.
			if len(values) == 0 {
				logger.Info("Auth >> Header missing",
					zap.String("header", k),
					zap.String("ctx", util.FiberContextString(p.C)),
				)

				return m.handleFallback()
			}

			// Any (pattern, value) match passes the header check.
			matched := false

			for _, v := range values {
				for _, re := range res {
					if re.MatchString(v) {
						matched = true

						break
					}
				}

				if matched {
					break
				}
			}

			if !matched {
				logger.Info("Auth >> Header value not matched",
					zap.String("header", k),
					zap.Any("patterns", m.auth.Header[k]),
					zap.Any("actual", values),
					zap.String("ctx", util.FiberContextString(p.C)),
				)

				return m.handleFallback()
			}
		}

		// Query checks
		for k, re := range m.queryRE {
			if !re.MatchString(p.C.Query(k)) {
				logger.Info("Auth >> Query value not matched",
					zap.String("key", k),
					zap.Any("pattern", m.auth.Query[k]),
					zap.String("actual", p.C.Query(k)),
					zap.String("ctx", util.FiberContextString(p.C)),
				)

				return m.handleFallback()
			}
		}

		return p, nil
	})
}

func (m *AuthModifier) initRegex() {
	// Pre-compile patterns so per-request does not hit regexp cache locks.
	if len(m.auth.Header) > 0 {
		m.headerRE = make(map[string][]*regexp.Regexp, len(m.auth.Header))
		for k, patterns := range m.auth.Header {
			out := make([]*regexp.Regexp, 0, len(patterns))
			for _, p := range patterns {
				pat, ok := reader.LiteralTrim(p)
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
			pat, ok := reader.LiteralTrim(p)
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

	wl := &ipAllowlist{ips: []net.IP{}, nets: []*net.IPNet{}}

	add := func(raw string) error {
		s := strings.TrimSpace(raw)
		if s == "" || strings.HasPrefix(s, "#") {
			return nil
		}

		if strings.Contains(s, "/") {
			_, n, err := net.ParseCIDR(s)
			if err != nil {
				return err
			}

			wl.nets = append(wl.nets, n)

			return nil
		}

		ip := net.ParseIP(s)
		if ip == nil {
			return errors.New("invalid ip")
		}

		wl.ips = append(wl.ips, ip)

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
			lineNo := i + 1

			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			if err := add(line); err != nil {
				m.ipErr = err
				if logger != nil {
					logger.Warn("Auth >> Invalid ipAllowlistFile entry",
						zap.Int("line", lineNo),
						zap.String("value", line),
						zap.Error(err),
					)
				}

				return
			}
		}
	}

	if len(wl.ips) == 0 && len(wl.nets) == 0 {
		m.ipWL = nil
		m.ipErr = nil

		return
	}

	m.ipWL = wl
}

func (wl *ipAllowlist) allowAny(ips []string) bool {
	for _, raw := range ips {
		ipStr := strings.TrimSpace(raw)
		if ipStr == "" {
			continue
		}

		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}

		for _, allowed := range wl.ips {
			if allowed.Equal(ip) {
				return true
			}
		}

		for _, n := range wl.nets {
			if n.Contains(ip) {
				return true
			}
		}
	}

	return false
}

func (m *AuthModifier) handleFallback() (*ProcessData, error) {
	if m.auth.Fallback == nil || m.auth.Fallback.Type == "" {
		return nil, &AuthFallbackNextError{}
	}

	switch m.auth.Fallback.Type {
	case config.AuthFallbackNext:
		return nil, &AuthFallbackNextError{}
	case config.AuthFallbackJump:
		return nil, &AuthFallbackJumpError{JumpTo: m.auth.Fallback.JumpTo}
	default:
		return nil, &AuthFallbackNextError{}
	}
}
