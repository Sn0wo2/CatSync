package response

func New(msg string, data ...any) *Response {
	switch len(data) {
	case 0:
		return &Response{
			Msg: msg,
		}
	case 1:
		return &Response{
			Msg:  msg,
			Data: data[0],
		}
	default:
		return &Response{
			Msg:  msg,
			Data: data,
		}
	}
}
