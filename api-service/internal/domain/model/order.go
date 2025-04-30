package model

import (
	"github.com/google/uuid"
)

type OrderItem struct {
	ProductID int64 `json:"product_id"`
	Quantity  int64 `json:"quantity"`
	Price     int64 `json:"price"`
}

type Order struct {
	ID         string      `json:"id"`
	CustomerID string      `json:"customer_id"`
	Items      []OrderItem `json:"items"`
	TotalPrice int64       `json:"total_price"`
	Status     string      `json:"status"`
	CreatedAt  string      `json:"created_at"`
	UpdatedAt  string      `json:"updated_at"`
}

type CreateOrderReq struct {
	CustomerID uuid.UUID   `json:"customer_id"`
	Items      []OrderItem `json:"items"`
}

type CreateOrderResp struct {
	Order *Order `json:"order"`
}
