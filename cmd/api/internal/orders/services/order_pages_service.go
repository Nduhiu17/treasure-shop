package services

import (
	"context"
	"time"

	"github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderPagesService struct {
	col *mongo.Collection
}

func NewOrderPagesService(db *mongo.Database) *OrderPagesService {
	return &OrderPagesService{col: db.Collection("order_pages")}
}

func (s *OrderPagesService) Create(pages *models.OrderPages) error {
	pages.ID = primitive.NewObjectID()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.col.InsertOne(ctx, pages)
	return err
}

func (s *OrderPagesService) List() ([]models.OrderPages, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := s.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var pages []models.OrderPages
	if err := cur.All(ctx, &pages); err != nil {
		return nil, err
	}
	return pages, nil
}

func (s *OrderPagesService) GetByID(id primitive.ObjectID) (*models.OrderPages, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var pages models.OrderPages
	err := s.col.FindOne(ctx, bson.M{"_id": id}).Decode(&pages)
	if err != nil {
		return nil, err
	}
	return &pages, nil
}

func (s *OrderPagesService) Update(id primitive.ObjectID, update bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.col.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func (s *OrderPagesService) Delete(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
