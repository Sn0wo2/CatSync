package action

type Modifier interface {
	ProcessModifier(handler Handler) Handler
}

type ModifiableHandler struct {
	Handler
	modifiers []Modifier
}

func WrapHandler(h Handler) *ModifiableHandler {
	return &ModifiableHandler{
		Handler:   h,
		modifiers: []Modifier{},
	}
}

func (h *ModifiableHandler) WithModifier(m Modifier) *ModifiableHandler {
	h.modifiers = append(h.modifiers, m)

	return h
}

func (h *ModifiableHandler) Build() Handler {
	result := h.Handler
	for _, m := range h.modifiers {
		result = m.ProcessModifier(result)
	}

	return result
}
