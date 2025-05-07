package proxy

import (
	"context"
	"fmt"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"order-microsystem/api-service/internal/domain/model"
	"order-microsystem/api-service/pkg/config"
	pb "order-microsystem/api-service/pkg/proto/inventory"
	"time"
)

type InventoryProxy struct {
	client pb.InventoryServiceClient
	conn   *grpc.ClientConn
}

func NewInventoryProxy(cfg *config.Config) (*InventoryProxy, error) {
	hystrix.ConfigureCommand("InventoryProxy", hystrix.CommandConfig{
		Timeout:                cfg.Hystrix.Timeout,
		MaxConcurrentRequests:  cfg.Hystrix.MaxConcurrentRequests,
		RequestVolumeThreshold: cfg.Hystrix.RequestVolumeThreshold,
		SleepWindow:            cfg.Hystrix.SleepWindow,
		ErrorPercentThreshold:  cfg.Hystrix.ErrorPercentThreshold,
	})

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

	services, _, err := client.Health().Service(cfg.Service.Inventory.Name, "", true, nil)
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

	return &InventoryProxy{
		client: pb.NewInventoryServiceClient(conn),
		conn:   conn,
	}, nil
}

func (p *InventoryProxy) GetAllInventory(ctx context.Context, offset int32, limit int32) ([]*model.Inventory, error) {
	var inventories []*model.Inventory

	err := hystrix.Do("InventoryService", func() error {
		resp, err := p.client.GetAllInventory(ctx, &pb.GetAllInventoryRequest{
			Offset: offset,
			Limit:  limit,
		})
		if err != nil {
			return err
		}

		for _, item := range resp.Products {
			inventories = append(inventories, &model.Inventory{
				ProductID:   item.ProductId,
				ProductName: item.ProductName,
				Price:       item.Price,
				Quantity:    item.Quantity,
			})
		}

		return nil
	}, func(err error) error {
		return errors.Errorf("fallback triggered due to error: %v", err)
	})

	if err != nil {
		return nil, err
	}

	return inventories, nil
}
