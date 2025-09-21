package entities

import "time"

type Order struct {
	ID         uint      `json:"id"`
	ProductID  uint      `json:"product_id"`
	Qty        int       `json:"qty"`
	TotalPrice float64   `json:"total_price,omitempty"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
}
