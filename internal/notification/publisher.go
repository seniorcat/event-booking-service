package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"laschool.ru/event-booking-service/internal/config"
)

type Publisher interface {
	Publish(ctx context.Context, event *NotificationEvent) error
	Close() error
}

type publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	cfg     config.RabbitMQ
}

func NewPublisher(cfg config.RabbitMQ) (Publisher, error) {
	conn, err := amqp.Dial(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	err = ch.ExchangeDeclare(
		cfg.Exchange,
		cfg.ExchangeType,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare an exchange: %w", err)
	}

	q, err := ch.QueueDeclare(
		cfg.QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	err = ch.QueueBind(
		q.Name,
		cfg.RoutingKey,
		cfg.Exchange,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind a queue: %w", err)
	}

	return &publisher{
		conn:    conn,
		channel: ch,
		cfg:     cfg,
	}, nil
}

func (p *publisher) Publish(ctx context.Context, event *NotificationEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		log.Printf("failed to marshal notification event: %v", err)
		return err
	}

	err = p.channel.PublishWithContext(
		ctx,
		p.cfg.Exchange,
		p.cfg.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		log.Printf("failed to publish notification event: %v", err)
		return err
	}

	log.Printf("Published notification: event=%s, booking_id=%d, user_id=%d",
		event.Event, event.BookingID, event.UserID)
	return nil
}

func (p *publisher) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}
