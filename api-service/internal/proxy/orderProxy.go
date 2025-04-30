package proxy

import (
	"context"
	"fmt"
	"github.com/hashicorp/consul/api"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"order-microsystem/api-service/internal/domain/model"
	"order-microsystem/api-service/pkg/config"
	pb "order-microsystem/api-service/pkg/proto/order"
	"time"
)

type OrderProxy struct {
	client pb.OrderServiceClient
	conn   *grpc.ClientConn
}

func NewOrderProxy(cfg *config.Config) (*OrderProxy, error) {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = cfg.Consul.Address

	var client *api.Client
	var err error

	// 添加重试逻辑
	for i := 0; i < 5; i++ {
		client, err = api.NewClient(consulConfig)
		if err == nil {
			break
		}
		time.Sleep(time.Second * time.Duration(i+1))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %v", err)
	}

	services, _, err := client.Health().Service(cfg.Service.Order.Name, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query service: %v", err)
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("no healthy instances available")
	}

	service := services[0].Service
	address := fmt.Sprintf("%s:%d", service.Address, service.Port)

	conn, err := grpc.NewClient(
		address,
		grpc.WithStatsHandler(otelgrpc.NewClientHandler(
			otelgrpc.WithTracerProvider(otel.GetTracerProvider()),
			otelgrpc.WithPropagators(otel.GetTextMapPropagator()),
		)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to grpc server: %v", err)
	}

	return &OrderProxy{
		client: pb.NewOrderServiceClient(conn),
		conn:   conn,
	}, nil
}

func (p *OrderProxy) CreateOrder(ctx *context.Context, order *model.CreateOrderReq) (*model.Order, error) {
	var items []*pb.OrderItem
	for _, item := range order.Items {
		items = append(items, &pb.OrderItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}
	req := &pb.CreateOrderRequest{
		CustomerId: order.CustomerID.String(),
		Items:      items,
	}

	resp, err := p.client.CreateOrder(*ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %v", err)
	}
	var orderItems []model.OrderItem
	for _, item := range resp.Order.Items {
		orderItems = append(orderItems, model.OrderItem{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}
	return &model.Order{
		ID:         resp.Order.Id,
		CustomerID: resp.Order.CustomerId,
		Items:      orderItems,
		TotalPrice: resp.Order.TotalPrice,
		Status:     resp.Order.Status,
		CreatedAt:  resp.Order.CreatedAt,
		UpdatedAt:  resp.Order.UpdatedAt,
	}, nil
}
