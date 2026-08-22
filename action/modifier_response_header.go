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

func (m *ResponseHeaderModifier) WithUpstream(upstream string) *ResponseHeaderModifier {
	m.upstream = upstream

	return m
}

func (m *ResponseHeaderModifier) Before(p *ProcessData) (*ProcessData, ExecutionResult) {
	var upstreamMap map[string][]string

	for k, v := range m.header {
		kLower := strings.ToLower(k)

		if len(v) == 1 && v[0] == "$[UPSTREAM_HEADER]" {
			if upstreamMap == nil {
				if m.upstream == "" {
					continue
				}

				resp, err := headerFetcher.Fetch(m.upstream)
				if err != nil {
					return nil, ExecutionResult{Err: fmt.Errorf("fetch upstream error: %w", err)}
				}

				upstreamMap = make(map[string][]string, len(resp.Header))
				for k, v := range resp.Header {
					upstreamMap[strings.ToLower(k)] = v
				}
			}

			for _, val := range upstreamMap[kLower] {
				p.FCtx.Append(k, val)
			}
		} else {
			p.FCtx.Append(k, v...)
		}
	}

	return p, ExecutionResult{}
}

func (m *ResponseHeaderModifier) After(p *ProcessData) (*ProcessData, ExecutionResult) {
	return p, ExecutionResult{}
}
