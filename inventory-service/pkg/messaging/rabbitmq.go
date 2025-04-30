package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"log"
	"order-microsystem/inventory-service/internal/domain/repository"
	"order-microsystem/inventory-service/pkg/config"
	"time"
)

type RabbitMQ struct {
	conn   *amqp091.Connection
	ch     *amqp091.Channel
	config *config.RabbitMQConfig
	repo   *repository.MySQLRepository
}

func NewRabbitMQ(config *config.RabbitMQConfig, repo *repository.MySQLRepository) (*RabbitMQ, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d", config.Username, config.Password, config.Host, config.Port)
	var conn *amqp091.Connection
	var err error
	for i := 0; i < 5; i++ {
		conn, err = amqp091.Dial(url)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to rabbitmq: %v", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %v", err)
	}

	// declare exchange
	err = channel.ExchangeDeclare(
		config.Exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %v", err)
	}

	return &RabbitMQ{
		conn:   conn,
		ch:     channel,
		config: config,
		repo:   repo,
	}, nil
}

func (rmq *RabbitMQ) Close() error {
	if err := rmq.ch.Close(); err != nil {
		return err
	}
	return rmq.conn.Close()
}

func (rmq *RabbitMQ) ConsumeOrderCreated() {
	// 声明一个匿名队列
	q, err := rmq.ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = rmq.ch.QueueBind(
		q.Name,
		"order.created",
		rmq.config.Exchange,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err.Error())
	}

	msgs, err := rmq.ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	for msg := range msgs {
		var receive_msg struct {
			EventType string    `json:"event_type"`
			OrderID   uuid.UUID `json:"order_id"`
			UserID    uuid.UUID `json:"user_id"`
			Status    string    `json:"status"`
			Products  []struct {
				ProductID int64 `json:"product_id"`
				Quantity  int64 `json:"quantity"`
				Price     int64 `json:"price"`
			} `json:"products"`
			TotalPrice int64  `json:"total_price"`
			CreatedAt  string `json:"created_at"`
		}
		if err := json.Unmarshal(msg.Body, &receive_msg); err != nil {
			log.Fatal(err.Error())
			msg.Nack(false, true)
			continue
		}

		for _, item := range receive_msg.Products {
			source, err := rmq.repo.GetInventory(item.ProductID)
			if err != nil {
				log.Fatalf("ProductID not exists: %v", err)
				continue
			}
			quantity := source.Quantity - item.Quantity
			err = rmq.repo.UpdateInventory(item.ProductID, quantity)
			if err != nil {
				log.Fatalf("Update inventory failed: %v", err)
				continue
			}
		}

		fmt.Println("totalprice: ", receive_msg.TotalPrice)
		rmq.PublishInventoryLocked(receive_msg.OrderID, receive_msg.UserID, receive_msg.TotalPrice)
		msg.Ack(false)
	}
}

func (rmq *RabbitMQ) PublishInventoryLocked(orderID uuid.UUID, userID uuid.UUID, totalPrice int64) error {
	event := map[string]interface{}{
		"event_type":  "inventory_locked",
		"order_id":    orderID,
		"user_id":     userID,
		"total_price": totalPrice,
	}

	// 序列化
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = rmq.ch.PublishWithContext(ctx,
		rmq.config.Exchange,
		"inventory.locked",
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	return nil
}
