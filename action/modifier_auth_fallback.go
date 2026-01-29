package action

import "fmt"

type AuthFallbackJumpError struct {
	JumpTo int
}

func (e *AuthFallbackJumpError) Error() string {
	return fmt.Sprintf("auth fallback: jump to %d", e.JumpTo)
}

type AuthFallbackNextError struct{}

func (e *AuthFallbackNextError) Error() string {
	return "auth fallback: next"
}
