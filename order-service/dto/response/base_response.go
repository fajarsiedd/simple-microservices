package response

type BaseResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
	Data    any    `json:"data"`
}