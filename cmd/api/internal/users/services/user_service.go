package services

import (
	"context"
	"errors"
	"time"

	"github.com/nduhiu17/treasure-shop/cmd/api/internal/users/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userCollection *mongo.Collection
}

func NewUserService(db *mongo.Database) *UserService {
	return &UserService{
		userCollection: db.Collection("users"),
	}
}

func (s *UserService) CreateUser(user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	existingUser := s.userCollection.FindOne(ctx, bson.M{"email": user.Email})
	if existingUser.Err() == nil {
		return errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	_, err = s.userCollection.InsertOne(ctx, user)
	return err
}

func (s *UserService) GetUserByID(id primitive.ObjectID) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := s.userCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	return &user, err
}

func (s *UserService) GetUsersByRole(role string) ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := s.userCollection.Find(ctx, bson.M{"roles": role})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (s *UserService) UpdateUser(user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{}
	if user.Email != "" {
		update["email"] = user.Email
	}
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		update["password"] = string(hashedPassword)
	}
	if len(user.Roles) > 0 {
		update["roles"] = user.Roles
	}
	// Add other fields you want to update

	_, err := s.userCollection.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": update})
	return err
}

func (s *UserService) DeleteUser(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.userCollection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
