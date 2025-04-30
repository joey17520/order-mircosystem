package repository

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log"
	"order-microsystem/inventory-service/internal/domain/model"
)

type MySQLRepository struct {
	db *gorm.DB
}

func NewMySQLRepository(db *gorm.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (m *MySQLRepository) AutoMigrations() error {
	err := m.db.AutoMigrate(&model.Product{})
	if err != nil {
		return fmt.Errorf("failed to autoMigrate Product model: %v", err)
	}
	var total int64
	m.db.Model(&model.Product{}).Count(&total)
	if total > 0 {
		return nil
	}
	preparedData := []model.Product{
		{ProductID: 1001, ProductName: "IPhone11", Price: 799999, Quantity: 9999},
		{ProductID: 1002, ProductName: "IPhone12", Price: 899999, Quantity: 9999},
		{ProductID: 1003, ProductName: "IPhone13", Price: 999999, Quantity: 9999},
	}
	if err := m.db.Create(&preparedData).Error; err != nil {
		log.Fatalf("failed to insert prepared data: %v", err)
	}
	return nil
}

func (m *MySQLRepository) GetInventory(productID int64) (*model.Product, error) {
	var product model.Product
	if err := m.db.Where("product_id = ?", productID).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	return &product, nil
}

func (m *MySQLRepository) GetAllInventory(offset int32, limit int32) ([]*model.Product, error) {
	var products []*model.Product

	result := m.db.Find(&products).Offset(int(offset)).Limit(int(limit))
	if result.Error != nil {
		return nil, result.Error
	}

	return products, nil
}

func (m *MySQLRepository) UpdateInventory(productID int64, quantity int64) error {
	result := m.db.Model(&model.Product{}).
		Where("product_id = ?", productID).
		Update("quantity", quantity)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
