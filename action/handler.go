package action

import (
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/params"
	"github.com/gofiber/fiber/v2"
)

type ProcessData struct {
	Ctx     *params.Ctx
	C       *fiber.Ctx
	PayLoad config.ActionData
}

type Handler interface {
	ProcessAction(data *ProcessData) error
	HookProcessData() func(*ProcessData) *ProcessData
}

type BaseHandler struct{}

func (h *BaseHandler) HookProcessData() func(*ProcessData) *ProcessData {
	return nil
}
