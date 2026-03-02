package action

import (
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/params"
	"github.com/gofiber/fiber/v3"
)

type ProcessData struct {
	Ctx     *params.Ctx
	C       fiber.Ctx
	Action  *config.Action
	PayLoad config.ActionData
}

type Handler interface {
	ProcessAction(data *ProcessData) error
}

type HookFunc func(*ProcessData) (*ProcessData, error)

type HandlerWithHooks struct {
	Handler
	beforeHooks []HookFunc
	afterHooks  []HookFunc
}

func WrapHandlerWithHooks(h Handler) *HandlerWithHooks {
	return &HandlerWithHooks{
		Handler:     h,
		beforeHooks: []HookFunc{},
		afterHooks:  []HookFunc{},
	}
}

func (h *HandlerWithHooks) Before(hook HookFunc) *HandlerWithHooks {
	h.beforeHooks = append(h.beforeHooks, hook)

	return h
}

func (h *HandlerWithHooks) After(hook HookFunc) *HandlerWithHooks {
	h.afterHooks = append(h.afterHooks, hook)

	return h
}

func (h *HandlerWithHooks) ProcessAction(data *ProcessData) error {
	var err error

	for _, hook := range h.beforeHooks {
		data, err = hook(data)
		if err != nil {
			return err
		}
	}

	err = h.Handler.ProcessAction(data)
	if err != nil {
		return err
	}

	for _, hook := range h.afterHooks {
		data, err = hook(data)
		if err != nil {
			return err
		}
	}

	return nil
}
