package service

import "order-microsystem/inventory-service/internal/domain/model"

type InventoryRepository interface {
	UpdateInventory(product_id int64, quantity int64) error
	GetInventory(product_id int64) (*model.Product, error)
	GetAllInventory(offset int32, limit int32) ([]*model.Product, error)
}

type InventoryService struct {
	repo InventoryRepository
}

func NewInventoryService(repo InventoryRepository) *InventoryService {
	return &InventoryService{
		repo: repo,
	}
}

func (s *InventoryService) GetInventory(productID int64) (*model.Product, error) {
	result, err := s.repo.GetInventory(productID)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *InventoryService) GetAllInventory(offset int32, limit int32) ([]*model.Product, error) {
	results, err := s.repo.GetAllInventory(offset, limit)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (s *InventoryService) UpdateInventory(productID int64, quantity int64) error {
	err := s.repo.UpdateInventory(productID, quantity)
	if err != nil {
		return err
	}
	return nil
}
