package action

type ResponseHeaderModifier struct {
	key    string
	values []string
}

func NewResponseHeaderModifier(key string, values ...string) *ResponseHeaderModifier {
	return &ResponseHeaderModifier{key: key, values: values}
}

func (m *ResponseHeaderModifier) ProcessModifier(handler Handler) Handler {
	return WrapHandlerWithHooks(handler).Before(func(p *ProcessData) (*ProcessData, error) {
		p.C.Append(m.key, m.values...)
		return p, nil
	})
}
