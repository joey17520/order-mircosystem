package service

import "order-microsystem/payment-service/internal/domain/model"

type PaymentRepository interface {
	CreatePayment(paymentModel *model.PaymentModel) error
	GetPayment(user_id string) (*model.PaymentModel, error)
	GetAllPayment(user_id string) ([]*model.PaymentModel, error)
}

type PaymentService struct {
	repo PaymentRepository
}

func NewPaymentService(repo PaymentRepository) *PaymentService {
	return &PaymentService{repo: repo}
}

func (s *PaymentService) CreatePayment(model *model.PaymentModel) error {
	if err := s.repo.CreatePayment(model); err != nil {
		return err
	}
	return nil
}

func (s *PaymentService) GetPayment(user_id string) (*model.PaymentModel, error) {
	result, err := s.repo.GetPayment(user_id)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *PaymentService) GetAllPayment(user_id string) ([]*model.PaymentModel, error) {
	results, err := s.repo.GetAllPayment(user_id)
	if err != nil {
		return nil, err
	}
	return results, nil
}
