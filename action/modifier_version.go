package action

import "strings"

type PlaceholderModifier struct {
	placeholder string
	value       string
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

func (p *PlaceholderModifier) Before(pd *ProcessData) (*ProcessData, ExecutionResult) {
	return pd, ExecutionResult{}
}

func (p *PlaceholderModifier) After(pd *ProcessData) (*ProcessData, ExecutionResult) {
	for k, vals := range pd.FCtx.GetRespHeaders() {
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

		pd.FCtx.Response().Header.Del(k)

		out := vals[:0]
		for _, vv := range vals {
			if strings.Contains(vv, p.placeholder) {
				out = append(out, strings.ReplaceAll(vv, p.placeholder, p.value))
			} else {
				out = append(out, vv)
			}
		}

		pd.FCtx.Append(k, out...)
	}

	return pd, ExecutionResult{}
}
