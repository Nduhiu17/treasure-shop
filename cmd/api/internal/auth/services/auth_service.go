package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/users/models"
	userservices "github.com/nduhiu17/treasure-shop/cmd/api/internal/users/services"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userCollection *mongo.Collection
}

func NewAuthService(db *mongo.Database) *AuthService {
	return &AuthService{
		userCollection: db.Collection("users"),
	}
}

func (s *AuthService) Register(user *models.User, userRoleService *userservices.UserRoleService, roleService *userservices.RoleService) error {
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

	// Generate random 6-digit user_number
	user.UserNumber = generateRandomSixDigitNumber()

	res, err := s.userCollection.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	// Find the 'user' role
	userRoleObj, err := roleService.GetByName("user")
	if err != nil {
		return errors.New("default user role not found")
	}

	// Create user_roles document
	userID := res.InsertedID.(primitive.ObjectID)
	userRoleDoc := &models.UserRole{
		UserID: userID,
		RoleID: userRoleObj.ID,
	}
	if err := userRoleService.Create(userRoleDoc); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) Login(email, password string) (string, *models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := s.userCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return "", nil, errors.New("invalid email address")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", nil, errors.New("invalid password")
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = user.ID.Hex()
	claims["email"] = user.Email
	claims["roles"] = []string{} // Initialize roles slice
	claims["user"] = user
	claims["user_number"] = user.UserNumber
	fmt.Println("JWT user_number:", user.UserNumber)

	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	claims["iat"] = time.Now().Unix()
	// Fetch user roles
	fmt.Println("user ID for roles:", user.UserNumber)
	userRoleService := userservices.NewUserRoleService(s.userCollection.Database())
	roles, err := userRoleService.GetByUserID(user.ID)
	if err != nil {
		return "", nil, errors.New("failed to fetch user roles")
	}
	var roleNames []string
	for _, role := range roles {
		roleService := userservices.NewRoleService(s.userCollection.Database())
		roleObj, err := roleService.GetByID(role.RoleID)
		if err != nil {
			return "", nil, errors.New("failed to fetch user roles")
		}
		roleNames = append(roleNames, roleObj.Name)
	}
	claims["roles"] = roleNames

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key" // Replace with a strong secret in .env
	}

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", nil, err
	}

	// Remove password before returning user
	user.Password = ""
	user.Roles = roleNames

	return tokenString, &user, nil
}

// generateRandomSixDigitNumber returns a random 6-digit string
func generateRandomSixDigitNumber() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}
