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

type VersionModifier struct {
	Version string
}

func NewVersionModifier(version string) *VersionModifier {
	return &VersionModifier{Version: version}
}

func (v *VersionModifier) ProcessModifier(handler Handler) Handler {
	return &wrappedHandler{
		Handler: handler,
		hook: func(p *ProcessData) *ProcessData {
			p.C.Append("X-Version", v.Version)

			return p
		},
	}
}
