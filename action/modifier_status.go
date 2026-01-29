package action

import "fmt"

type StatusModifier struct {
	status uint16
}

func NewStatusModifier() *StatusModifier {
	return &StatusModifier{}
}

func (v *StatusModifier) WithStatus(status uint16) *StatusModifier {
	v.status = status

	return v
}

func (v *StatusModifier) Status() uint16 {
	return v.status
}

func (v *StatusModifier) ProcessModifier(handler Handler) Handler {
	return WrapHandlerWithHooks(handler).Before(func(p *ProcessData) (*ProcessData, error) {
		if v.status == 0 {
			return p, nil
		}

		if v.status < 100 || v.status > 599 {
			return nil, fmt.Errorf("invalid status code: %d", v.status)
		}

		p.C.Status(int(v.status))

		return p, nil
	})
}
