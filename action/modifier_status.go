package action

import (
	"fmt"
	"net/http"
	"time"
)

type StatusModifier struct {
	status   uint16
	upstream string
}

func NewStatusModifier() *StatusModifier {
	return &StatusModifier{}
}

func (v *StatusModifier) WithStatus(status uint16) *StatusModifier {
	v.status = status

	return v
}

func (v *StatusModifier) WithUpstream(upstream string) *StatusModifier {
	v.upstream = upstream

	return v
}

func (v *StatusModifier) Before(p *ProcessData) (*ProcessData, ExecutionResult) {
	status := v.status
	if v.upstream != "" {
		client := &http.Client{Timeout: 10 * time.Second}

		resp, err := client.Get(v.upstream)
		if err != nil {
			return nil, ExecutionResult{Err: err}
		}

		_ = resp.Body.Close()

		if resp.StatusCode < 100 || resp.StatusCode > 599 {
			return nil, ExecutionResult{Err: fmt.Errorf("invalid upstream status code: %d", resp.StatusCode)}
		}

		status = uint16(resp.StatusCode)
	}

	if status < 100 || status > 599 {
		return nil, ExecutionResult{Err: fmt.Errorf("invalid status code: %d", status)}
	}

	p.C.Status(int(status))

	return p, ExecutionResult{}
}

func (v *StatusModifier) After(p *ProcessData) (*ProcessData, ExecutionResult) {
	return p, ExecutionResult{}
}
