package services

import (
	"context"
	"time"

	"github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderUrgencyService struct {
	col *mongo.Collection
}

func NewOrderUrgencyService(db *mongo.Database) *OrderUrgencyService {
	return &OrderUrgencyService{col: db.Collection("order_urgency")}
}

func (s *OrderUrgencyService) Create(urgency *models.OrderUrgency) error {
	urgency.ID = primitive.NewObjectID()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.col.InsertOne(ctx, urgency)
	return err
}

func (s *OrderUrgencyService) List() ([]models.OrderUrgency, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := s.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var urgencies []models.OrderUrgency
	if err := cur.All(ctx, &urgencies); err != nil {
		return nil, err
	}
	return urgencies, nil
}

func (s *OrderUrgencyService) GetByID(id primitive.ObjectID) (*models.OrderUrgency, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var urgency models.OrderUrgency
	err := s.col.FindOne(ctx, bson.M{"_id": id}).Decode(&urgency)
	if err != nil {
		return nil, err
	}
	return &urgency, nil
}

func (s *OrderUrgencyService) Update(id primitive.ObjectID, update bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.col.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func (s *OrderUrgencyService) Delete(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
