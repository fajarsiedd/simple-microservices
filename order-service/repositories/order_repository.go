package repositories

import "order-service/entities"

type OrderRepository interface {
	Create(order entities.Order) (entities.Order, error)
	FindAll() ([]entities.Order, error)
	FindByID(id uint) (entities.Order, error)
	FindByProductID(productID uint) ([]entities.Order, error)
	Update(order entities.Order) (entities.Order, error)
	Delete(id uint) error
}
