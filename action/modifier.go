package action

type Modifier interface {
	Before(data *ProcessData) (*ProcessData, ExecutionResult)
	After(data *ProcessData) (*ProcessData, ExecutionResult)
}

type ModifiableHandler struct {
	Handler
	modifiers []Modifier
}

func WrapHandler(h Handler) *ModifiableHandler {
	return &ModifiableHandler{
		Handler: h,
	}
}

func (h *ModifiableHandler) WithModifier(m Modifier) *ModifiableHandler {
	h.modifiers = append(h.modifiers, m)

	return h
}

func (h *ModifiableHandler) ProcessAction(data *ProcessData) ExecutionResult {
	for i := len(h.modifiers) - 1; i >= 0; i-- {
		var result ExecutionResult

		data, result = h.modifiers[i].Before(data)
		if result.Err != nil || result.Status != ExecutionCompleted {
			return result
		}
	}

	if result := h.Handler.ProcessAction(data); result.Err != nil || result.Status != ExecutionCompleted {
		return result
	}

	for _, modifier := range h.modifiers {
		var result ExecutionResult

		data, result = modifier.After(data)
		if result.Err != nil || result.Status != ExecutionCompleted {
			return result
		}
	}

	return ExecutionResult{}
}
