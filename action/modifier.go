package action

type Modifier interface {
	ProcessModifier(handler Handler) Handler
}

type wrappedHandler struct {
	Handler
	hook func(*ProcessData) *ProcessData
}

func (h *wrappedHandler) HookProcessData() func(*ProcessData) *ProcessData {
	return h.hook
}
