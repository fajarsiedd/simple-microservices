package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"order-service/database"
	"order-service/entities"
	"order-service/messaging"
	"order-service/repositories"
	"os"
	"strconv"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

const cacheTTL = 5 * time.Minute

type Product struct {
	ID    uint    `json:"id"`
	Price float64 `json:"price"`
	Qty   int     `json:"qty"`
}

type ProductResponse struct {
	Data Product
}

type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
}

type orderService struct {
	orderRepo  repositories.OrderRepository
	httpClient HTTPClient
	messaging  messaging.MessagingService
	cache      database.CacheService
	exchange   string
}

func NewOrderService(
	orderRepo repositories.OrderRepository,
	httpClient HTTPClient,
	messaging messaging.MessagingService,
	cache database.CacheService,
) OrderService {
	return &orderService{
		orderRepo:  orderRepo,
		httpClient: httpClient,
		messaging:  messaging,
		cache:      cache,
		exchange:   os.Getenv("RABBITMQ_EXCHANGE_NAME"),
	}
}

func (s *orderService) Create(order entities.Order) (entities.Order, error) {
	if order.ProductID == 0 || order.Qty <= 0 {
		return entities.Order{}, fmt.Errorf("invalid order data")
	}

	go func(order entities.Order) {
		eventPayload := map[string]interface{}{
			"productID": order.ProductID,
			"qty":       order.Qty,
		}

		jsonData, err := json.Marshal(eventPayload)
		if err != nil {
			log.Printf("Failed to marshal order payload: %v", err)
			return
		}

		if err := s.messaging.PublishEvent(
			s.exchange,
			"order.created.request",
			jsonData,
		); err != nil {
			log.Printf("Failed to publish RabbitMQ event: %v", err)
		}
	}(order)

	return order, nil
}

func (s *orderService) StartOrderConsumer() {
	ch, err := s.messaging.GetChannel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	err = ch.Qos(
		10,    // prefetchCount
		0,     // prefetchSize
		false, // global
	)
	if err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
	}

	msgs, err := s.messaging.Consume("order-service.order.requests", "order.created.request")
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	log.Println("Order consumer started, waiting for messages...")

	for d := range msgs {
		s.processMessage(d)
	}
}

func (s *orderService) processMessage(d amqp091.Delivery) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ERROR: Consumer panicked while processing message: %v", r)
			if err := d.Nack(false, true); err != nil {
				log.Printf("Failed to Nack message on panic: %v", err)
			}
		}
	}()

	log.Printf("Received a message from queue: %s", string(d.Body))

	var orderRequest map[string]interface{}
	if err := json.Unmarshal(d.Body, &orderRequest); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		d.Ack(false) // Acknowledge and drop invalid message
		return
	}

	productID := uint(orderRequest["productID"].(float64))
	qty := int(orderRequest["qty"].(float64))

	productURL := fmt.Sprintf("%s/products/%d", os.Getenv("PRODUCT_SERVICE_URL"), productID)
	resp, err := s.httpClient.Get(productURL)
	if err != nil {
		log.Printf("Failed to call product-service: %v", err)
		d.Nack(false, true) // Requeue
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Product not found or invalid response status: %d", resp.StatusCode)

		eventPayload := map[string]interface{}{
			"productID": productID,
			"qty":       qty,
			"reason":    "Product not found",
			"timestamp": time.Now(),
		}
		jsonData, _ := json.Marshal(eventPayload)
		s.messaging.PublishEvent(os.Getenv("RABBITMQ_EXCHANGE_NAME"), "order.failed", jsonData)

		d.Ack(false)
		return
	}

	var productResp ProductResponse
	if err := json.NewDecoder(resp.Body).Decode(&productResp); err != nil {
		log.Printf("Failed to decode product data: %v", err)

		d.Nack(false, false) // Nack without requeue
		return
	}

	if productResp.Data.Qty < qty {
		log.Printf("Insufficient stock for product ID: %d", productID)

		eventPayload := map[string]interface{}{
			"productID": productID,
			"qty":       qty,
			"reason":    "Insufficient stock",
			"timestamp": time.Now(),
		}
		jsonData, _ := json.Marshal(eventPayload)
		s.messaging.PublishEvent(os.Getenv("RABBITMQ_EXCHANGE_NAME"), "order.failed", jsonData)

		d.Ack(false)
		return
	}

	order := entities.Order{
		ProductID:  productID,
		Qty:        qty,
		Status:     "completed",
		TotalPrice: productResp.Data.Price * float64(qty),
	}

	createdOrder, err := s.orderRepo.Create(order)
	if err != nil {
		log.Printf("Failed to create order in DB: %v", err)
		d.Nack(false, true) // Requeue
		return
	}

	s.cache.Del("orders:id:" + strconv.Itoa(int(createdOrder.ID)))
	s.cache.Del("orders:productid:" + strconv.Itoa(int(createdOrder.ProductID)))

	eventPayload := map[string]interface{}{
		"pattern": "order.created",
		"data": map[string]interface{}{
			"orderID":   createdOrder.ID,
			"productID": createdOrder.ProductID,
			"qty":       createdOrder.Qty,
		},
	}
	jsonData, _ := json.Marshal(eventPayload)
	err = s.messaging.PublishEvent(os.Getenv("RABBITMQ_EXCHANGE_NAME"), "order.created", jsonData)
	if err != nil {
		log.Printf("Failed to publish RabbitMQ event: %v", err)
	}

	d.Ack(false) // Acknowledge message after successful processing
	log.Printf("Successfully processed order ID: %d", createdOrder.ID)
}

func (s *orderService) StartOrderFailedConsumer() {
    msgs, err := s.messaging.SetupFailedQueue()
    if err != nil {
        log.Fatalf("Failed to setup failed queue: %v", err)
    }

    log.Println("Failed order consumer started, waiting for messages...")

    for d := range msgs {
        log.Printf("Received failed order: %s", string(d.Body))
        d.Ack(false)
    }
}

// func (s *orderService) Create(order entities.Order) (entities.Order, error) {
// 	productURL := fmt.Sprintf("%s/products/%d", os.Getenv("PRODUCT_SERVICE_URL"), order.ProductID)
// 	resp, err := s.httpClient.Get(productURL) // Gunakan s.httpClient
// 	if err != nil {
// 		return entities.Order{}, fmt.Errorf("failed to call product-service: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return entities.Order{}, fmt.Errorf("product not found with ID: %d", order.ProductID)
// 	}

// 	var product ProductResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
// 		return entities.Order{}, fmt.Errorf("failed to decode product data: %w", err)
// 	}

// 	if product.Data.Qty < order.Qty {
// 		return entities.Order{}, fmt.Errorf("insufficient stock")
// 	}

// 	order.TotalPrice = product.Data.Price * float64(order.Qty)

// 	createdOrder, err := s.orderRepo.Create(order)
// 	if err != nil {
// 		return entities.Order{}, err
// 	}

// 	s.cache.Del("orders:id:" + strconv.Itoa(int(createdOrder.ID)))
// 	s.cache.Del("orders:productid:" + strconv.Itoa(int(createdOrder.ProductID)))

// 	eventPayload := map[string]interface{}{
// 		"pattern": "order.created",
// 		"data":    createdOrder,
// 	}

// 	jsonData, _ := json.Marshal(eventPayload)
// 	// Gunakan s.messaging
// 	err = s.messaging.PublishEvent(os.Getenv("RABBITMQ_EXCHANGE_NAME"), "order.created", jsonData)
// 	if err != nil {
// 		log.Printf("Failed to publish RabbitMQ event: %v", err)
// 	}

// 	return createdOrder, nil
// }

func (s *orderService) FindAll() ([]entities.Order, error) {
	return s.orderRepo.FindAll()
}

func (s *orderService) FindByID(id uint) (entities.Order, error) {
	cacheKey := fmt.Sprintf("orders:id:%d", id)
	val, err := s.cache.Get(cacheKey)
	if err == nil && val != "" {
		var order entities.Order
		json.Unmarshal([]byte(val), &order)
		return order, nil
	}

	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return entities.Order{}, err
	}

	jsonData, _ := json.Marshal(order)
	s.cache.SetWithTTL(cacheKey, string(jsonData), cacheTTL)
	return order, nil
}

func (s *orderService) FindByProductID(productID uint) ([]entities.Order, error) {
	cacheKey := fmt.Sprintf("orders:productid:%d", productID)
	val, err := s.cache.Get(cacheKey)
	if err == nil && val != "" {
		var orders []entities.Order
		json.Unmarshal([]byte(val), &orders)
		return orders, nil
	}

	orders, err := s.orderRepo.FindByProductID(productID)
	if err != nil {
		return nil, err
	}

	jsonData, _ := json.Marshal(orders)
	s.cache.SetWithTTL(cacheKey, string(jsonData), cacheTTL)
	return orders, nil
}

func (s *orderService) Update(order entities.Order) (entities.Order, error) {
	updatedOrder, err := s.orderRepo.Update(order)
	if err != nil {
		return entities.Order{}, err
	}

	s.cache.Del("orders:id:" + strconv.Itoa(int(updatedOrder.ID)))
	s.cache.Del("orders:productid:" + strconv.Itoa(int(updatedOrder.ProductID)))

	return updatedOrder, nil
}

func (s *orderService) Delete(id uint) error {
	s.cache.Del("orders:id:" + strconv.Itoa(int(id)))

	return s.orderRepo.Delete(id)
}
