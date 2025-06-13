package services

import (
	"context"
	"time"

	"github.com/nduhiu17/treasure-shop/cmd/api/internal/users/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRoleService struct {
	col *mongo.Collection
}

func NewUserRoleService(db *mongo.Database) *UserRoleService {
	return &UserRoleService{col: db.Collection("user_roles")}
}

func (s *UserRoleService) Create(userRole *models.UserRole) error {
	userRole.ID = primitive.NewObjectID()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.col.InsertOne(ctx, userRole)
	return err
}

func (s *UserRoleService) GetByUserID(userID primitive.ObjectID) ([]models.UserRole, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := s.col.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var userRoles []models.UserRole
	for cur.Next(ctx) {
		var ur models.UserRole
		if err := cur.Decode(&ur); err != nil {
			return nil, err
		}
		userRoles = append(userRoles, ur)
	}
	return userRoles, nil
}

func (s *UserRoleService) List() ([]models.UserRole, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := s.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var userRoles []models.UserRole
	for cur.Next(ctx) {
		var ur models.UserRole
		if err := cur.Decode(&ur); err != nil {
			return nil, err
		}
		userRoles = append(userRoles, ur)
	}
	return userRoles, nil
}

func (s *UserRoleService) Delete(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
