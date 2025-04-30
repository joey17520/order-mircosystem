package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"order-microsystem/order-service/internal/domain/model"
	"order-microsystem/order-service/pkg/cache"
	"order-microsystem/order-service/pkg/messaging"
	"time"
)

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id string) (*model.Order, error)
	UpdateStatus(ctx context.Context, id string, status model.OrderStatus) error
}

type OrderService struct {
	repo        OrderRepository
	rabbitmq    *messaging.RabbitMQ
	redisClient *cache.RedisClient
}

func NewOrderService(repo OrderRepository, rabbitmq *messaging.RabbitMQ, redis *cache.RedisClient) *OrderService {
	return &OrderService{
		repo:        repo,
		rabbitmq:    rabbitmq,
		redisClient: redis,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, customerID uuid.UUID, items []model.OrderItem) (*model.Order, error) {
	// 计算总价
	totalPrice := int64(0)
	for _, item := range items {
		totalPrice += item.Price * item.Quantity
	}

	order := &model.Order{
		ID:         uuid.New(),
		UserID:     customerID,
		Items:      items,
		TotalPrice: totalPrice,
		Status:     "pending",
		CreatedAt:  time.Now().Format(time.RFC3339),
		UpdatedAt:  time.Now().Format(time.RFC3339),
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %v", err)
	}

	if err := s.rabbitmq.PublishOrderCreated(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to publish order created event: %v", err)
	}

	if err := s.redisClient.Set(fmt.Sprintf("order_%s", order.ID), order); err != nil {
		return nil, fmt.Errorf("failed to set order in cache: %v", err)
	}
	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (*model.Order, error) {
	// 先从缓存中获取订单
	var jsonOorder model.Order
	err := s.redisClient.Get(fmt.Sprintf("order_%s", id), &jsonOorder)
	if err == nil {
		return &jsonOorder, nil
	}
	return s.repo.GetByID(ctx, id)
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, id string, status model.OrderStatus) (*model.Order, error) {
	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		return nil, err
	}
	updatedOrder, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// 更新缓存
	if err := s.redisClient.Set(fmt.Sprintf("order_%s", id), updatedOrder); err != nil {
		return nil, fmt.Errorf("failed to update order in cache: %v", err)
	}
	// 获取更新后的订单
	return updatedOrder, err
}
