package action

type Type int

const (
	File         = iota
	String       = 1
	TempRedirect = 2
	Redirect     = 3
)
