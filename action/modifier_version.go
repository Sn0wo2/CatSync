package action

import "strings"

type PlaceholderModifier struct {
	placeholder string
	value       string
	header      bool
}

func NewPlaceholderModifier() *PlaceholderModifier {
	return &PlaceholderModifier{}
}

func (p *PlaceholderModifier) WithPlaceholder(placeholder string) *PlaceholderModifier {
	p.placeholder = placeholder

	return p
}

func (p *PlaceholderModifier) WithValue(value string) *PlaceholderModifier {
	p.value = value

	return p
}

func (p *PlaceholderModifier) WithHeader(header bool) *PlaceholderModifier {
	p.header = header

	return p
}

func (p *PlaceholderModifier) Value() string {
	return p.value
}

func (p *PlaceholderModifier) ProcessModifier(handler Handler) Handler {
	// Run in After hook so response headers are already populated
	// by other modifiers and the action handler.
	return WrapHandlerWithHooks(handler).After(func(pd *ProcessData) (*ProcessData, error) {
		for k, vals := range pd.C.GetRespHeaders() {
			keyHas := p.header && strings.Contains(k, p.placeholder)
			if !keyHas {
				found := false

				for _, vv := range vals {
					if strings.Contains(vv, p.placeholder) {
						found = true

						break
					}
				}

				if !found {
					continue
				}
			}

			pd.C.Response().Header.Del(k)

			newKey := k
			if p.header && strings.Contains(k, p.placeholder) {
				newKey = strings.ReplaceAll(k, p.placeholder, p.value)
			}

			out := vals[:0]
			for _, vv := range vals {
				if strings.Contains(vv, p.placeholder) {
					out = append(out, strings.ReplaceAll(vv, p.placeholder, p.value))
				} else {
					out = append(out, vv)
				}
			}

			if len(out) > 0 {
				pd.C.Append(newKey, out...)
			}
		}

		return pd, nil
	})
}
