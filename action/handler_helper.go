package action

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
	return nil
}
