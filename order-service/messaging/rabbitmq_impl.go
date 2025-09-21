package messaging

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type messagingService struct {
	conn *amqp091.Connection
	mu   sync.Mutex
}

func NewRabbitMQService() *messagingService {
	return &messagingService{}
}

func (s *messagingService) ConnectRabbitMQ() error {
	amqpURI := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		os.Getenv("RABBITMQ_USER"),
		os.Getenv("RABBITMQ_PASSWORD"),
		os.Getenv("RABBITMQ_HOST"),
		os.Getenv("RABBITMQ_PORT"),
	)

	var err error
	s.conn, err = amqp091.Dial(amqpURI)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	log.Println("RabbitMQ connection successfully opened!")
	return nil
}

func (s *messagingService) GetChannel() (*amqp091.Channel, error) {
	if s.conn == nil {
		return nil, fmt.Errorf("RabbitMQ connection is not established")
	}

	ch, err := s.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return ch, nil
}

func (s *messagingService) PublishEvent(exchangeName, routingKey string, body []byte) error {
	if s.conn == nil {
		return fmt.Errorf("RabbitMQ connection is not established")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ch, err := s.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return ch.PublishWithContext(ctx,
		exchangeName,
		routingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

func (s *messagingService) Consume(queueName string, routingKey string) (<-chan amqp091.Delivery, error) {
	if s.conn == nil {
		return nil, fmt.Errorf("RabbitMQ connection is not established")
	}

	ch, err := s.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	err = ch.QueueBind(
		q.Name,
		routingKey,
		os.Getenv("RABBITMQ_EXCHANGE_NAME"),
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bind queue to exchange: %w", err)
	}

	return ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
}

func (s *messagingService) SetupFailedQueue() (<-chan amqp091.Delivery, error) {
	if s.conn == nil {
		return nil, fmt.Errorf("RabbitMQ connection is not established")
	}

	ch, err := s.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	q, err := ch.QueueDeclare(
		"order-service.order.failed", // queue name
		true,                         // durable
		false,                        // auto-delete
		false,                        // exclusive
		false,                        // no-wait
		nil,                          // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	err = ch.QueueBind(
		q.Name,
		"order.failed",
		os.Getenv("RABBITMQ_EXCHANGE_NAME"),
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	return ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
}
