package action

import "fmt"

type ErrAuthFallbackJump struct {
	JumpTo int
}

func (e *ErrAuthFallbackJump) Error() string {
	return fmt.Sprintf("auth fallback: jump to %d", e.JumpTo)
}

type ErrAuthFallbackNext struct{}

func (e *ErrAuthFallbackNext) Error() string {
	return "auth fallback: next"
}
