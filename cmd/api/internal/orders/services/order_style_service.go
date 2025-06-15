package services

import (
	"context"
	"time"

	"github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderStyleService struct {
	col *mongo.Collection
}

func NewOrderStyleService(db *mongo.Database) *OrderStyleService {
	return &OrderStyleService{col: db.Collection("order_style")}
}

func (s *OrderStyleService) Create(style *models.OrderStyle) error {
	style.ID = primitive.NewObjectID()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.col.InsertOne(ctx, style)
	return err
}

func (s *OrderStyleService) List() ([]models.OrderStyle, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := s.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var styles []models.OrderStyle
	if err := cur.All(ctx, &styles); err != nil {
		return nil, err
	}
	return styles, nil
}

func (s *OrderStyleService) GetByID(id primitive.ObjectID) (*models.OrderStyle, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var style models.OrderStyle
	err := s.col.FindOne(ctx, bson.M{"_id": id}).Decode(&style)
	if err != nil {
		return nil, err
	}
	return &style, nil
}

func (s *OrderStyleService) Update(id primitive.ObjectID, update bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.col.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func (s *OrderStyleService) Delete(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
