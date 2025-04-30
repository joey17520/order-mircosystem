package model

import "github.com/google/uuid"

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

type OrderItem struct {
	ProductID int64 `json:"product_id" bson:"product_id"`
	Quantity  int64 `json:"quantity" bson:"quantity"`
	Price     int64 `json:"price" bson:"price"`
}

type Order struct {
	ID         uuid.UUID   `json:"id" bson:"_id,omitempty"`
	UserID     uuid.UUID   `json:"user_id" bson:"user_id"`
	Items      []OrderItem `json:"items" bson:"items"`
	TotalPrice int64       `json:"total_price" bson:"total_price"`
	Status     OrderStatus `json:"status" bson:"status"`
	CreatedAt  string      `json:"created_at" bson:"created_at"`
	UpdatedAt  string      `json:"updated_at" bson:"updated_at"`
}
