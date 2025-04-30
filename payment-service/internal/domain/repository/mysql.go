package repository

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"order-microsystem/payment-service/internal/domain/model"
)

type MySQLRepository struct {
	db *gorm.DB
}

func NewMySQLRepository(db *gorm.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) AutoMigration() error {
	if err := r.db.AutoMigrate(&model.PaymentModel{}); err != nil {
		return fmt.Errorf("failed to autoMigrate Product model: %v", err)
	}
	return nil
}

func (r *MySQLRepository) CreatePayment(paymentModel *model.PaymentModel) error {
	if err := r.db.Create(paymentModel).Error; err != nil {
		return fmt.Errorf("failed to create payment: %v", err)
	}
	return nil
}

func (r *MySQLRepository) GetPayment(user_id string) (*model.PaymentModel, error) {
	var payment model.PaymentModel
	if err := r.db.Where("user_id = ?", user_id).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	return &payment, nil
}

func (r *MySQLRepository) GetAllPayment(user_id string) ([]*model.PaymentModel, error) {
	var payments []*model.PaymentModel
	if err := r.db.Where("user_id = ?", user_id).Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}
