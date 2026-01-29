package action

import (
	"github.com/Sn0wo2/CatSync/config"
	util "github.com/Sn0wo2/CatSync/internal/util"
	"go.uber.org/zap"
	"strings"
)

type AuthModifier struct {
	auth config.ActionModifierAuth
}

func NewAuthModifier(auth config.ActionModifierAuth) *AuthModifier {
	return &AuthModifier{auth: auth}
}

func (m *AuthModifier) ProcessModifier(handler Handler) Handler {
	return WrapHandlerWithHooks(handler).Before(func(p *ProcessData) (*ProcessData, error) {
		logger := p.Ctx.GetLogger()

		// Header checks
		reqHeaders := p.C.GetReqHeaders()
		for k, patterns := range m.auth.Header {
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
				return m.handleFallback(p)
			}

			// Any (pattern, value) match passes the header check.
			matched := false
			for _, v := range values {
				for _, pattern := range patterns {
					re, err := util.GetCompiledRegexp(pattern)
					if err != nil {
						logger.Warn("Auth >> Invalid header regexp",
							zap.String("pattern", pattern),
							zap.Error(err),
						)
						continue
					}
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
					zap.Any("patterns", patterns),
					zap.Any("actual", values),
					zap.String("ctx", util.FiberContextString(p.C)),
				)
				return m.handleFallback(p)
			}
		}

		// Query checks
		for k, v := range m.auth.Query {
			re, err := util.GetCompiledRegexp(v)
			if err != nil {
				logger.Warn("Auth >> Invalid query regexp",
					zap.String("pattern", v),
					zap.Error(err),
				)
				continue
			}

			if !re.MatchString(p.C.Query(k)) {
				logger.Info("Auth >> Query value not matched",
					zap.String("key", k),
					zap.String("pattern", v),
					zap.String("actual", p.C.Query(k)),
					zap.String("ctx", util.FiberContextString(p.C)),
				)

				return m.handleFallback(p)
			}
		}

		return p, nil
	})
}

func (m *AuthModifier) handleFallback(p *ProcessData) (*ProcessData, error) {
	if m.auth.Fallback == nil || m.auth.Fallback.Type == "" {
		return nil, &ErrAuthFallbackNext{}
	}

	switch m.auth.Fallback.Type {
	case config.AuthFallbackNext:
		return nil, &ErrAuthFallbackNext{}
	case config.AuthFallbackJump:
		return nil, &ErrAuthFallbackJump{JumpTo: int(m.auth.Fallback.JumpTo)}
	default:
		return nil, &ErrAuthFallbackNext{}
	}
}
