package action

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Sn0wo2/CatSync/internal/upstream"
)

type ResponseHeaderModifier struct {
	header   http.Header
	upstream string
}

var headerFetcher = upstream.NewFetcher(30 * time.Second)

func (m *ResponseHeaderModifier) Before(p *ProcessData) (*ProcessData, ExecutionResult) {
	var (
		upstreamHeaders http.Header
		upstreamFetched bool
	)

	for k, v := range m.header {
		if len(v) == 1 && v[0] == "$[UPSTREAM_HEADER]" {
			if !upstreamFetched {
				if m.upstream == "" {
					continue
				}

				resp, err := headerFetcher.Fetch(m.upstream)
				if err != nil {
					return nil, ExecutionResult{Err: fmt.Errorf("fetch upstream error: %w", err)}
				}

				upstreamHeaders = resp.Header
				upstreamFetched = true
			}

			for _, val := range upstreamHeaders.Values(k) {
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
