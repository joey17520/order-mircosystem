package controller

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"order-microsystem/inventory-service/internal/service"
	pb "order-microsystem/inventory-service/pkg/proto/inventory"
)

type InventoryController struct {
	pb.UnimplementedInventoryServiceServer
	svc *service.InventoryService
}

func NewInventoryController(svc *service.InventoryService) *InventoryController {
	return &InventoryController{
		svc: svc,
	}
}

func RegisterInventoryController(server *grpc.Server, svc *InventoryController) {
	pb.RegisterInventoryServiceServer(server, svc)
}

func (c *InventoryController) GetAllInventory(ctx context.Context, req *pb.GetAllInventoryRequest) (*pb.GetAllInventoryResponse, error) {
	resp, err := c.svc.GetAllInventory(req.Offset, req.Limit)
	if err != nil {
		log.Fatalf("GetAllInventory failed: %v", err)
		return nil, err
	}
	var products []*pb.Product
	for _, item := range resp {
		products = append(products, &pb.Product{
			ProductId:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
		})
	}
	return &pb.GetAllInventoryResponse{Products: products}, nil
}
