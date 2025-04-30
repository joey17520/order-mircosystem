package controller

import (
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"log"
	"order-microsystem/order-service/internal/domain/model"
	"order-microsystem/order-service/internal/service"
	pb "order-microsystem/order-service/pkg/proto/order"
)

type OrderController struct {
	pb.UnimplementedOrderServiceServer
	svc *service.OrderService
}

func NewOrderController(svc *service.OrderService) *OrderController {
	return &OrderController{svc: svc}
}

func RegisterOrderService(server *grpc.Server, svc *OrderController) {
	pb.RegisterOrderServiceServer(server, svc)
}

func (s *OrderController) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	// 转换请求参数为领域模型
	items := make([]model.OrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, model.OrderItem{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}

	// 调用服务层
	createdOrder, err := s.svc.CreateOrder(ctx, convertToUUID(req.CustomerId), items)
	if err != nil {
		return nil, err
	}

	// 转换响应
	return &pb.CreateOrderResponse{
		Order: &pb.Order{
			Id:         createdOrder.ID.String(),
			CustomerId: createdOrder.UserID.String(),
			Items:      req.Items, // 保持原始items
			TotalPrice: createdOrder.TotalPrice,
			Status:     string(createdOrder.Status),
			CreatedAt:  createdOrder.CreatedAt,
			UpdatedAt:  createdOrder.UpdatedAt,
		},
	}, nil
}

func (s *OrderController) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	order, err := s.svc.GetOrder(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	// 转换响应
	return &pb.GetOrderResponse{
		Order: &pb.Order{
			Id:         order.ID.String(),
			CustomerId: order.UserID.String(),
			Items:      convertToProtoItems(order.Items),
			TotalPrice: order.TotalPrice,
			Status:     string(order.Status),
			CreatedAt:  order.CreatedAt,
			UpdatedAt:  order.UpdatedAt,
		},
	}, nil
}

func (s *OrderController) UpdateOrder(ctx context.Context, req *pb.UpdateOrderRequest) (*pb.UpdateOrderResponse, error) {
	updatedOrder, err := s.svc.UpdateOrderStatus(ctx, req.Id, model.OrderStatus(req.Status))
	if err != nil {
		return nil, err
	}

	// 转换响应
	return &pb.UpdateOrderResponse{
		Order: &pb.Order{
			Id:         updatedOrder.ID.String(),
			CustomerId: updatedOrder.UserID.String(),
			Items:      convertToProtoItems(updatedOrder.Items),
			TotalPrice: updatedOrder.TotalPrice,
			Status:     string(updatedOrder.Status),
			CreatedAt:  updatedOrder.CreatedAt,
			UpdatedAt:  updatedOrder.UpdatedAt,
		},
	}, nil
}

// 辅助函数：转换领域模型到proto消息
func convertToProtoItems(items []model.OrderItem) []*pb.OrderItem {
	protoItems := make([]*pb.OrderItem, 0, len(items))
	for _, item := range items {
		protoItems = append(protoItems, &pb.OrderItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}
	return protoItems
}

func convertToUUID(key string) uuid.UUID {
	uid, err := uuid.Parse(key)
	if err != nil {
		log.Fatal("Can't parse string to UUID")
	}
	return uid
}
