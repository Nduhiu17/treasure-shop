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

func (s *UserService) CreateUser(user *models.User, userRoleService *UserRoleService, roleService *RoleService, roleID ...primitive.ObjectID) error {
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

	res, err := s.userCollection.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	var assignedRole *models.Role
	if len(roleID) > 0 {
		// Use provided roleID
		assignedRole = &models.Role{ID: roleID[0]}
	} else {
		// Find the 'user' role
		assignedRole, err = roleService.GetByName("user")
		if err != nil {
			return errors.New("default user role not found")
		}
	}

	// Create user_roles document
	userID := res.InsertedID.(primitive.ObjectID)
	userRoleDoc := &models.UserRole{
		UserID: userID,
		RoleID: assignedRole.ID,
	}
	if err := userRoleService.Create(userRoleDoc); err != nil {
		return err
	}

	return nil
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
	// Remove roles update logic since roles are now managed via user_roles

	_, err := s.userCollection.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": update})
	return err
}

func (s *UserService) DeleteUser(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.userCollection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
