package action

type Data = any

type Type int

const (
	File = iota
	String
	TempRedirect
	Redirect
	JSON
)
