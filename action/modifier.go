package action

type Modifier interface {
	ProcessModifier(handler Handler) Handler
}

type wrappedHandler struct {
	Handler
	hook func(*ProcessData) (*ProcessData, error)
}

func (h *wrappedHandler) HookProcessData() func(*ProcessData) (*ProcessData, error) {
	return h.hook
}
