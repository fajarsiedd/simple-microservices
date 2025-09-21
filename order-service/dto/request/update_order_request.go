package request

type UpdateOrderRequest struct {
	Status string `json:"status" validate:"required"`
}
