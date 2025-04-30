package mongodb

import (
	"context"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"order-microsystem/order-service/internal/domain/model"
	"time"
)

type OrderRepository struct {
	collection *mongo.Collection
}

func NewOrderRepository(db *mongo.Database) *OrderRepository {
	return &OrderRepository{
		collection: db.Collection("orders"),
	}
}

func (r *OrderRepository) Create(ctx context.Context, order *model.Order) error {
	_, err := r.collection.InsertOne(ctx, bson.M{
		"_id":         order.ID.String(),
		"user_id":     order.UserID.String(),
		"items":       order.Items,
		"total_price": order.TotalPrice,
		"status":      order.Status,
		"created_at":  order.CreatedAt,
		"updated_at":  order.UpdatedAt,
	})
	return err
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*model.Order, error) {
	var result struct {
		ID         string            `bson:"_id"`
		UserID     string            `bson:"user_id"`
		Items      []model.OrderItem `bson:"items"`
		TotalPrice int64             `bson:"total_price"`
		Status     model.OrderStatus `bson:"status"`
		CreatedAt  string            `bson:"created_at"`
		UpdatedAt  string            `bson:"updated_at"`
	}

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &model.Order{
		ID:         convertUUID(result.ID),
		UserID:     convertUUID(result.UserID),
		Items:      result.Items,
		TotalPrice: result.TotalPrice,
		Status:     result.Status,
		CreatedAt:  result.CreatedAt,
		UpdatedAt:  result.UpdatedAt,
	}, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status model.OrderStatus) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"status":     status,
				"updated_at": time.Now().Format(time.RFC3339),
			},
		},
	)
	return err
}

func convertUUID(source string) uuid.UUID {
	parse, err := uuid.Parse(source)
	if err != nil {
		log.Fatalf("UUID parse failed: %v", err)
	}
	return parse
}
