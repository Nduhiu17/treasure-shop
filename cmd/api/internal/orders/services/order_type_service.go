package services

import (
	"context"
	"time"

	"github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderTypeService struct {
	Collection *mongo.Collection
}

func NewOrderTypeService(col *mongo.Collection) *OrderTypeService {
	return &OrderTypeService{Collection: col}
}

func (s *OrderTypeService) Create(ctx context.Context, orderType *models.OrderType) error {
	orderType.ID = primitive.NewObjectID()
	orderType.CreatedAt = time.Now()
	orderType.UpdatedAt = time.Now()
	_, err := s.Collection.InsertOne(ctx, orderType)
	return err
}

func (s *OrderTypeService) GetByID(ctx context.Context, id primitive.ObjectID) (*models.OrderType, error) {
	var orderType models.OrderType
	err := s.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&orderType)
	if err != nil {
		return nil, err
	}
	return &orderType, nil
}

func (s *OrderTypeService) List(ctx context.Context) ([]models.OrderType, error) {
	cur, err := s.Collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var orderTypes []models.OrderType
	for cur.Next(ctx) {
		var ot models.OrderType
		if err := cur.Decode(&ot); err != nil {
			return nil, err
		}
		orderTypes = append(orderTypes, ot)
	}
	return orderTypes, nil
}

// ListPaginated returns paginated order types and total count
func (s *OrderTypeService) ListPaginated(ctx context.Context, page, pageSize int) ([]models.OrderType, int64, error) {
	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)
	total, err := s.Collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}
	cur, err := s.Collection.Find(ctx, bson.M{}, nil)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)
	var allOrderTypes []models.OrderType
	for cur.Next(ctx) {
		var ot models.OrderType
		if err := cur.Decode(&ot); err != nil {
			return nil, 0, err
		}
		allOrderTypes = append(allOrderTypes, ot)
	}
	start := int(skip)
	end := start + int(limit)
	if start > len(allOrderTypes) {
		return []models.OrderType{}, total, nil
	}
	if end > len(allOrderTypes) {
		end = len(allOrderTypes)
	}
	return allOrderTypes[start:end], total, nil
}

func (s *OrderTypeService) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	update["updated_at"] = time.Now()
	_, err := s.Collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func (s *OrderTypeService) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := s.Collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
