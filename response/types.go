package response

type Response struct {
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}
