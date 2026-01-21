package action

import (
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/params"
	"github.com/gofiber/fiber/v2"
)

type ProcessData struct {
	Ctx     *params.Ctx
	C       *fiber.Ctx
	Action  *config.Action
	PayLoad *config.ActionData
	Hooks   *[]Modifier
}

type Handler interface {
	ProcessAction(data *ProcessData) error
	HookProcessData() func(*ProcessData) (*ProcessData, error)
}

type modifierHandler struct {
	Handler
	hook func(*ProcessData) (*ProcessData, error)
}

func (h *modifierHandler) HookProcessData() func(*ProcessData) (*ProcessData, error) {
	return h.hook
}
