package messaging

import "github.com/rabbitmq/amqp091-go"

type MessagingService interface {
	ConnectRabbitMQ() error
	PublishEvent(exchangeName, routingKey string, body []byte) error
	Consume(queueName string, routingKey string) (<-chan amqp091.Delivery, error)
	GetChannel() (*amqp091.Channel, error)
	SetupFailedQueue() (<-chan amqp091.Delivery, error)
}
