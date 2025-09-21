package models

import (
	"order-service/entities"
	"time"
)

type Order struct {
	ID         uint      `gorm:"primaryKey"`
	ProductID  uint      `json:"product_id"`
	Qty        int       `json:"qty"`
	TotalPrice float64   `json:"total_price"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Orders []Order

func (o Order) FromEntity(order entities.Order) Order {
	return Order{
		ID:         order.ID,
		ProductID:  order.ProductID,
		TotalPrice: order.TotalPrice,
		Qty:        order.Qty,
		Status:     order.Status,
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
	}
}

func (o *Order) ToEntity() entities.Order {
	return entities.Order{
		ID:         o.ID,
		ProductID:  o.ProductID,
		Qty:        o.Qty,
		TotalPrice: o.TotalPrice,
		Status:     o.Status,
		CreatedAt:  o.CreatedAt,
		UpdatedAt:  o.UpdatedAt,
	}
}

func (os Orders) FromEntities(orders []entities.Order) Orders {
	data := Orders{}

	for _, v := range orders {
		data = append(data, Order{}.FromEntity(v))
	}

	return data
}

func (os *Orders) ToEntities() []entities.Order {
	data := []entities.Order{}

	for _, v := range *os {
		data = append(data, v.ToEntity())
	}

	return data
}
