package request

type CreateOrderRequest struct {
	ProductID uint `json:"product_id" validate:"required"`
	Qty       int  `json:"qty" validate:"required,gt=0"`
}
