package repositories

import (
	"order-service/entities"
	"order-service/models"

	"gorm.io/gorm"
)

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{
		db: db,
	}
}

func (r *orderRepository) Create(order entities.Order) (entities.Order, error) {
	orderModel := models.Order{}.FromEntity(order)

	if err := r.db.Create(&orderModel).Error; err != nil {
		return entities.Order{}, err
	}

	return orderModel.ToEntity(), nil
}

func (r *orderRepository) FindAll() ([]entities.Order, error) {
	var ordersModel models.Orders

	if err := r.db.Find(&ordersModel).Error; err != nil {
		return nil, err
	}

	return ordersModel.ToEntities(), nil
}

func (r *orderRepository) FindByID(id uint) (entities.Order, error) {
	orderModel := models.Order{}

	if err := r.db.First(&orderModel, &id).Error; err != nil {
		return entities.Order{}, err
	}

	return orderModel.ToEntity(), nil
}

func (r *orderRepository) FindByProductID(productID uint) ([]entities.Order, error) {
	var ordersModel models.Orders

	if err := r.db.Where("product_id = ?", productID).Find(&ordersModel).Error; err != nil {
		return nil, err
	}

	return ordersModel.ToEntities(), nil
}

func (r *orderRepository) Update(order entities.Order) (entities.Order, error) {
	orderModel := models.Order{}.FromEntity(order)

	if err := r.db.Updates(&orderModel).Error; err != nil {
		return entities.Order{}, err
	}

	if err := r.db.First(&orderModel, order.ID).Error; err != nil {
		return entities.Order{}, err
	}

	return orderModel.ToEntity(), nil
}

func (r *orderRepository) Delete(id uint) error {
	orderModel := models.Order{}

	if err := r.db.Delete(&orderModel, &id).Error; err != nil {
		return err
	}

	return nil
}
