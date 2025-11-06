package action

type Data = any

// Deprecated: Use action.Operation instead.
type Type int

const (
	File = iota
	String
	TempRedirect
	Redirect
	JSON
)

func (t *Type) ToOperation() Operation {
	switch *t {
	case String:
		return OperationString
	case TempRedirect:
		return OperationTempRedirect
	case Redirect:
		return OperationRedirect
	case JSON:
		return OperationJSON
	default: // 0(default): File
		return OperationFile
	}
}

type Operation string

const (
	OperationFile         Operation = "file"
	OperationString       Operation = "string"
	OperationTempRedirect Operation = "temp_redirect"
	OperationRedirect     Operation = "redirect"
	OperationJSON         Operation = "json"
)
