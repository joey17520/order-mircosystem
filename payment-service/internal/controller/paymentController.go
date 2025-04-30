package controller

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"order-microsystem/payment-service/internal/service"
	pb "order-microsystem/payment-service/pkg/proto/payment"
)

type PaymentController struct {
	pb.UnimplementedPaymentServiceServer
	svc *service.PaymentService
}

func NewPaymentController(svc *service.PaymentService) *PaymentController {
	return &PaymentController{
		svc: svc,
	}
}

func RegisterPaymentController(server *grpc.Server, svc *PaymentController) {
	pb.RegisterPaymentServiceServer(server, svc)
}

func (c *PaymentController) GetPayment(ctx context.Context, req *pb.GetPaymentRequest) (*pb.GetPaymentResponse, error) {
	payment, err := c.svc.GetPayment(req.UserId)
	if err != nil {
		log.Fatalf("GetPayment failed: %v", err)
	}
	return &pb.GetPaymentResponse{
		Payment: &pb.Payment{
			UserId:     payment.UserID.String(),
			OrderId:    payment.OrderID.String(),
			PaymentId:  payment.PaymentID.String(),
			TotalPrice: payment.TotalPrice,
		},
	}, nil
}

func (c *PaymentController) GetAllPayment(ctx context.Context, req *pb.GetAllPaymentRequest) (*pb.GetAllPaymentResponse, error) {
	payments, err := c.svc.GetAllPayment(req.UserId)
	if err != nil {
		log.Fatalf("GetAllPayment failed: %v", err)
	}
	var paymentsResp []*pb.Payment
	for _, payment := range payments {
		paymentsResp = append(paymentsResp, &pb.Payment{
			UserId:     payment.UserID.String(),
			OrderId:    payment.OrderID.String(),
			PaymentId:  payment.PaymentID.String(),
			TotalPrice: payment.TotalPrice,
		})
	}
	return &pb.GetAllPaymentResponse{
		Payments: paymentsResp,
	}, nil
}
