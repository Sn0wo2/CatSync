package action

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Sn0wo2/CatSync/internal/upstream"
)

type ResponseHeaderModifier struct {
	header   http.Header
	upstream string
}

var headerFetcher = upstream.NewFetcher(30 * time.Second)

func NewResponseHeaderModifier() *ResponseHeaderModifier {
	return &ResponseHeaderModifier{}
}

func (m *ResponseHeaderModifier) WithHeader(header http.Header) *ResponseHeaderModifier {
	m.header = header

	return m
}

func (m *ResponseHeaderModifier) WithUpStream(upstream string) *ResponseHeaderModifier {
	m.upstream = upstream

	return m
}

func (m *ResponseHeaderModifier) ProcessModifier(handler Handler) Handler {
	return WrapHandlerWithHooks(handler).Before(func(p *ProcessData) (*ProcessData, error) {
		hasUpstreamPlaceholder := false

		for _, v := range m.header {
			if len(v) == 1 && v[0] == "$[UPSTREAM_HEADER]" {
				hasUpstreamPlaceholder = true

				break
			}
		}

		var upstreamMap map[string][]string

		if m.upstream != "" && hasUpstreamPlaceholder {
			resp, err := headerFetcher.Fetch(m.upstream)
			if err != nil {
				return nil, fmt.Errorf("fetch upstream error: %w", err)
			}

			upstreamMap = make(map[string][]string, len(resp.Header))
			for k, v := range resp.Header {
				upstreamMap[strings.ToLower(k)] = v
			}
		}

		for k, v := range m.header {
			kLower := strings.ToLower(k)

			if len(v) == 1 && v[0] == "$[UPSTREAM_HEADER]" {
				if vals, ok := upstreamMap[kLower]; ok {
					for _, val := range vals {
						p.C.Append(k, val)
					}
				}
			} else {
				p.C.Append(k, v...)
			}
		}

		return p, nil
	})
}
