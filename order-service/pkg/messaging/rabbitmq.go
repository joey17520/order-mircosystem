package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"log"
	"order-microsystem/order-service/internal/domain/model"
	"order-microsystem/order-service/internal/domain/repository/mongodb"
	"order-microsystem/order-service/pkg/config"
	"time"
)

type RabbitMQ struct {
	conn   *amqp091.Connection
	ch     *amqp091.Channel
	config *config.RabbitMQConfig
	repo   *mongodb.OrderRepository
}

func NewRabbitMQ(config *config.RabbitMQConfig, repo *mongodb.OrderRepository) (*RabbitMQ, error) {
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

func (rmq *RabbitMQ) PublishOrderCreated(ctx context.Context, order *model.Order) error {
	// 构建订单创建事件的消息体
	event := map[string]interface{}{
		"event_type":  "order_created", // 事件类型标识
		"order_id":    order.ID,        // 订单唯一ID
		"user_id":     order.UserID,    // 下单用户ID
		"status":      order.Status,    // 订单状态
		"products":    order.Items,
		"total_price": order.TotalPrice,
		"created_at":  order.CreatedAt, // 订单创建时间
	}

	// 将事件对象序列化为JSON格式
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %v", err)
	}

	// 设置5秒超时上下文，防止消息发布阻塞
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel() // 确保在函数退出时取消上下文

	// 发布消息到RabbitMQ
	err = rmq.ch.PublishWithContext(ctx,
		rmq.config.Exchange, // 使用配置中指定的交换机
		"order.created",     // 路由键(routing key)
		false,               // 不强制(mandatory)
		false,               // 不立即(immediate)
		amqp091.Publishing{
			ContentType: "application/json", // 消息内容类型
			Body:        body,               // 消息体(JSON格式)
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	return nil
}

func (rmq *RabbitMQ) ConsumePaymentCompleted() {
	// 声明一个匿名队列
	q, err := rmq.ch.QueueDeclare(
		"",    // 队列名称为空，RabbitMQ会自动生成一个唯一的名称
		false, // 非持久化
		false, // 不自动删除
		true,  // 排他性，只对当前连接可见
		false, // 不等待
		nil,   // 额外参数
	)
	if err != nil {
		log.Fatal(err.Error())
	}

	// 绑定队列到交换机，指定路由键
	err = rmq.ch.QueueBind(
		q.Name,              // 队列名称
		"payment.completed", // 路由键
		rmq.config.Exchange, // 交换机名称
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err.Error())
	}

	// 消费消息
	msgs, err := rmq.ch.Consume(
		q.Name, // 队列名称
		"",     // 消费者标识
		false,  // 手动确认
		false,  // 非排他性
		false,  // 不等待
		false,  // 额外参数
		nil,    // 额外参数
	)

	for msg := range msgs {
		var event struct {
			EventType string    `json:"event_type"`
			UserID    uuid.UUID `json:"user_id"`
			OrderID   uuid.UUID `json:"order_id"`
		}
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Fatal(err.Error())
			msg.Nack(false, true) // 重试
			continue
		}

		// 更新订单状态
		if err := rmq.repo.UpdateStatus(context.Background(), event.OrderID.String(), model.OrderStatusCompleted); err != nil {
			log.Fatalf(err.Error())
			msg.Nack(false, true) //重试
			continue
		}
		msg.Ack(false) // 确认消息
	}
}
