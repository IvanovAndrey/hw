// internal/rmq/rmq.go

package rmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

type Notification struct {
	EventID  string    `json:"event_id"`
	Title    string    `json:"title"`
	DateTime time.Time `json:"datetime"`
	UserID   string    `json:"user_id"`
}

type Publisher interface {
	PublishNotification(ctx context.Context, note Notification) error
	Close() error
}

type Consumer interface {
	ConsumeNotifications(ctx context.Context, handleFunc func(Notification)) error
	Close() error
}

type RMQClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

func NewRMQClient(uri, queueName string) (*RMQClient, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("could not connect to RMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("could not open channel: %w", err)
	}

	queue, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("could not declare queue: %w", err)
	}

	return &RMQClient{conn: conn, channel: ch, queue: queue}, nil
}

func (r *RMQClient) PublishNotification(ctx context.Context, note Notification) error {
	body, err := json.Marshal(note)
	if err != nil {
		return fmt.Errorf("could not marshal notification: %w", err)
	}

	return r.channel.Publish(
		"", // exchange
		r.queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

func (r *RMQClient) ConsumeNotifications(ctx context.Context, handleFunc func(Notification)) error {
	msgs, err := r.channel.Consume(
		r.queue.Name,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-msgs:
				if !ok {
					return
				}
				var note Notification
				if err := json.Unmarshal(d.Body, &note); err != nil {
					log.Printf("could not decode message: %v", err)
					continue
				}
				handleFunc(note)
			}
		}
	}()

	<-ctx.Done()
	return nil
}

func (r *RMQClient) Close() error {
	if err := r.channel.Close(); err != nil {
		return err
	}
	return r.conn.Close()
}
