// order-service/main.go
package main

import (
	"log"
	"net/http"
	"time"

	"order-service/database"
	"order-service/handlers"
	"order-service/messaging"
	"order-service/middlewares"
	"order-service/repositories"
	"order-service/services"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, _ := database.Connect()

	cacheService := database.NewRedisService()

	msgService := messaging.NewRabbitMQService()
	if err := msgService.ConnectRabbitMQ(); err != nil {
		log.Fatalf("Could not connect to RabbitMQ: %v", err)
	}

	e := echo.New()

	// CORS
	e.Use(middleware.CORS())

	// Logger
	loggerConfig := middlewares.LoggerConfig{
		Format: `[${time_rfc3339}] ${status} ${method} ${host}${path} ${latency_human}` + "\n",
	}
	loggerMiddleware := loggerConfig.Init()
	e.Use(loggerMiddleware)
	e.Use(middleware.Recover())

	// Validator
	customValidator := middlewares.InitValidator()
	e.Validator = customValidator
	e.Pre(middleware.RemoveTrailingSlash())

	// Init routes
	repo := repositories.NewOrderRepository(db)
	service := services.NewOrderService(repo, &http.Client{Timeout: 10 * time.Second}, msgService, cacheService)
	handler := handlers.NewOrderHandler(service)

	order := e.Group("/orders")
	order.POST("", handler.CreateOrder)
	order.GET("", handler.FindAllOrders)
	order.GET("/:id", handler.FindOrderByID)
	order.GET("/product/:productID", handler.FindOrdersByProductID)
	order.PUT("/:id", handler.UpdateOrder)
	order.DELETE("/:id", handler.DeleteOrder)

	go service.StartOrderConsumer()
	go service.StartOrderFailedConsumer()

	e.Logger.Fatal(e.Start(":8080"))
}
