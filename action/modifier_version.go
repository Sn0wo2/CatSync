package action

type VersionModifier struct {
	version string
	header  bool
}

func NewVersionModifier() *VersionModifier {
	return &VersionModifier{}
}

func (v *VersionModifier) WithVersion(version string) *VersionModifier {
	v.version = version

	return v
}

func (v *VersionModifier) WithHeader(header bool) *VersionModifier {
	v.header = header

	return v
}

func (v *VersionModifier) Version() string {
	return v.version
}

func (v *VersionModifier) ProcessModifier(handler Handler) Handler {
	return &wrappedHandler{
		Handler: handler,
		hook: func(p *ProcessData) *ProcessData {
			p.C.Append("X-version", v.version)

			return p
		},
	}
}
