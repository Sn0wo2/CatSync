package action

import (
	"fmt"
	"net/http"
)

type StatusModifier struct {
	status uint16
	upstream string
}

func NewStatusModifier() *StatusModifier {
	return &StatusModifier{}
}

func (v *StatusModifier) WithStatus(status uint16) *StatusModifier {
	v.status = status

	return v
}

// WithUpstream 会覆盖掉 WithStatus
func (v *StatusModifier) WithUpstream(upstream string) *StatusModifier {
	v.upstream = upstream
	return v
}

func (v *StatusModifier) Status() (uint16, error) {
	if v.upstream != "" {
		resp, err := http.Get(v.upstream)
		if err != nil {
			return v.status, err
		}
		defer resp.Body.Close()
		return uint16(resp.StatusCode), nil
	}
	return v.status, nil
}

func (v *StatusModifier) ProcessModifier(handler Handler) Handler {
	return WrapHandlerWithHooks(handler).Before(func(p *ProcessData) (*ProcessData, error) {
		status, err := v.Status()
		if err != nil {
			return nil, err
		}

		if status == 0 || status < 100 || status > 599 {
			return nil, fmt.Errorf("invalid status code: %d", v.status)
		}

		p.C.Status(int(status))

		return p, nil
	})
}
