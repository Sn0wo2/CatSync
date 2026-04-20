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

func (v *StatusModifier) ProcessModifier(handler Handler) Handler {
	return WrapHandlerWithHooks(handler).Before(func(p *ProcessData) (*ProcessData, error) {
		status := v.status
		if v.upstream != "" {
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Get(v.upstream)
			if err != nil {
				return nil, err
			}

			statusCode := resp.StatusCode
			_ = resp.Body.Close()

			if statusCode < 100 || statusCode > 599 {
				return nil, fmt.Errorf("invalid upstream status code: %d", statusCode)
			}

			status = uint16(statusCode)
		}

		if status < 100 || status > 599 {
			return nil, fmt.Errorf("invalid status code: %d", status)
		}

		p.C.Status(int(status))

		return p, nil
	})
}
