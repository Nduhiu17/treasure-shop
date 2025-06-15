package services

import (
	"context"
	"time"

	"github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderLevelService struct {
	col *mongo.Collection
}

func NewOrderLevelService(db *mongo.Database) *OrderLevelService {
	return &OrderLevelService{col: db.Collection("order_levels")}
}

func (s *OrderLevelService) Create(level *models.OrderLevel) error {
	level.ID = primitive.NewObjectID()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.col.InsertOne(ctx, level)
	return err
}

func (s *OrderLevelService) List() ([]models.OrderLevel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := s.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var levels []models.OrderLevel
	if err := cur.All(ctx, &levels); err != nil {
		return nil, err
	}
	return levels, nil
}

func (s *OrderLevelService) GetByID(id primitive.ObjectID) (*models.OrderLevel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var level models.OrderLevel
	err := s.col.FindOne(ctx, bson.M{"_id": id}).Decode(&level)
	if err != nil {
		return nil, err
	}
	return &level, nil
}

func (s *OrderLevelService) Update(id primitive.ObjectID, update bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.col.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func (s *OrderLevelService) Delete(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
