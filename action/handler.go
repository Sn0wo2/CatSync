package action

import (
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/params"
	"github.com/gofiber/fiber/v3"
)

type ProcessData struct {
	Ctx      *params.Ctx
	C        fiber.Ctx
	Action   *config.Action
	Payload  config.ActionData
	Reloader Reloader
}

type Reloader interface {
	Reload() error
}

type Handler interface {
	ProcessAction(data *ProcessData) ExecutionResult
}

type ExecutionStatus uint8

const (
	ExecutionCompleted    ExecutionStatus = 0
	ExecutionFallbackNext ExecutionStatus = 2
	ExecutionFallbackJump ExecutionStatus = 3
)

type ExecutionResult struct {
	Status ExecutionStatus
	Err    error
	JumpTo int
}
