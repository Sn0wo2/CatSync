package action

import "fmt"

type Helper struct {
	processData *ProcessData
}

func NewHelper() *Helper {
	return &Helper{}
}

func (h *Helper) WithProcessData(data *ProcessData) *Helper {
	h.processData = data
	return h
}

func (h *Helper) GetProcessData() *ProcessData {
	return h.processData
}

func (h *Helper) ProcessStatus() error {
	if h.processData.Action.Status >= 100 && h.processData.Action.Status <= 599 {
		h.processData.C.Status(int(h.processData.Action.Status))
	} else if h.processData.Action.Status != 0 {
		return fmt.Errorf("invalid status code: %d", h.processData.Action.Status)
	}
	return nil
}
