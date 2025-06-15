package services

import (
	"context"
	"time"

	"github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderLanguageService struct {
	col *mongo.Collection
}

func NewOrderLanguageService(db *mongo.Database) *OrderLanguageService {
	return &OrderLanguageService{col: db.Collection("order_language")}
}

func (s *OrderLanguageService) Create(language *models.OrderLanguage) error {
	language.ID = primitive.NewObjectID()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.col.InsertOne(ctx, language)
	return err
}

func (s *OrderLanguageService) List() ([]models.OrderLanguage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := s.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var languages []models.OrderLanguage
	if err := cur.All(ctx, &languages); err != nil {
		return nil, err
	}
	return languages, nil
}

func (s *OrderLanguageService) GetByID(id primitive.ObjectID) (*models.OrderLanguage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var language models.OrderLanguage
	err := s.col.FindOne(ctx, bson.M{"_id": id}).Decode(&language)
	if err != nil {
		return nil, err
	}
	return &language, nil
}

func (s *OrderLanguageService) Update(id primitive.ObjectID, update bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.col.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func (s *OrderLanguageService) Delete(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
