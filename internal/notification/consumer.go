package notification

import (
	"context"
	"encoding/json"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"laschool.ru/event-booking-service/internal/config"
)

type Consumer interface {
	Consume(ctx context.Context) error
	Close() error
}

type consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	cfg     config.RabbitMQ
}

func NewConsumer(cfg config.RabbitMQ) (Consumer, error) {
	conn, err := amqp.Dial(cfg.DSN)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &consumer{
		conn:    conn,
		channel: ch,
		cfg:     cfg,
	}, nil
}

func (c *consumer) Consume(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		c.cfg.QueueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	log.Printf("Consumer started, listening on queue: %s", c.cfg.QueueName)

	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer received shutdown signal")
			return ctx.Err()

		case msg, ok := <-msgs:
			if !ok {
				log.Println("Message channel closed")
				return nil
			}

			if err := c.handleMessage(msg); err != nil {
				log.Printf("Error handling message: %v", err)
				msg.Nack(false, false)
			} else {
				msg.Ack(false)
			}
		}
	}
}

func (c *consumer) handleMessage(msg amqp.Delivery) error {
	var event NotificationEvent

	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("Failed to unmarshal message: %v, body: %s", err, string(msg.Body))
		return err
	}

	log.Printf("📧 EMAIL NOTIFICATION: Event type=%s, BookingID=%d, UserID=%d, EventID=%d, OccurredAt=%s",
		event.Event,
		event.BookingID,
		event.UserID,
		event.EventID,
		event.OccurredAt.Format(time.RFC3339),
	)

	return nil
}

func (c *consumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
