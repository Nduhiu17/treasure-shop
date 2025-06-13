package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderService struct {
	orderCollection *mongo.Collection
	userCollection  *mongo.Collection // For checking user/writer existence
}

func NewOrderService(db *mongo.Database) *OrderService {
	return &OrderService{
		orderCollection: db.Collection("orders"),
		userCollection:  db.Collection("users"),
	}
}

func (s *OrderService) CreateOrder(order *models.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	fmt.Println("Creating order:", order)
	_, err := s.orderCollection.InsertOne(ctx, order)
	return err
}

func (s *OrderService) GetAllOrders() ([]models.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cursor, err := s.orderCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *OrderService) GetOrdersByUserID(userID primitive.ObjectID) ([]models.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cursor, err := s.orderCollection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *OrderService) GetOrdersByStatus(status string) ([]models.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cursor, err := s.orderCollection.Find(ctx, bson.M{"status": status})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *OrderService) AssignOrder(orderID, writerID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify writer exists (multi-role system: check user_roles for writer role)
	writerRole := s.userCollection.Database().Collection("user_roles")
	roleCol := s.userCollection.Database().Collection("roles")
	var writerRoleIDs []primitive.ObjectID
	cursor, err := writerRole.Find(ctx, bson.M{"user_id": writerID})
	if err != nil {
		return errors.New("failed to check writer roles")
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var ur struct {
			RoleID primitive.ObjectID `bson:"role_id"`
		}
		if err := cursor.Decode(&ur); err == nil {
			writerRoleIDs = append(writerRoleIDs, ur.RoleID)
		}
	}
	isWriter := false
	for _, roleID := range writerRoleIDs {
		var roleDoc struct {
			Name string `bson:"name"`
		}
		if err := roleCol.FindOne(ctx, bson.M{"_id": roleID}).Decode(&roleDoc); err == nil && roleDoc.Name == "writer" {
			isWriter = true
			break
		}
	}
	if !isWriter {
		return errors.New("writer not found")
	}

	_, err = s.orderCollection.UpdateOne(
		ctx,
		bson.M{"_id": orderID, "status": "paid"},
		bson.M{"$set": bson.M{"writer_id": writerID, "status": "awaiting_asign_acceptance", "assignment_date": time.Now()}},
	)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("order already assigned to a writer")
		}
		if err == mongo.ErrNoDocuments {
			return errors.New("order not found or not awaiting assignment")
		}
	}
	return err
}

func (s *OrderService) SubmitOrder(orderID primitive.ObjectID, writerID primitive.ObjectID, content string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify order belongs to the writer and is in the correct status
	result := s.orderCollection.FindOne(ctx, bson.M{"_id": orderID, "writer_id": writerID, "status": "assigned"})
	if result.Err() != nil {
		return errors.New("order not found or not assigned to this writer")
	}

	_, err := s.orderCollection.UpdateOne(
		ctx,
		bson.M{"_id": orderID},
		bson.M{"$set": bson.M{"status": "submitted_for_review", "submission_date": time.Now()}},
	)
	if err != nil {
		return err
	}

	// In a real application, you would likely create a 'Review' record here as well.
	// For this basic example, we'll just update the order status.

	return nil
}

func (s *OrderService) ApproveOrder(orderID, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify order belongs to the user and is in the correct status
	result := s.orderCollection.FindOne(ctx, bson.M{"_id": orderID, "user_id": userID, "status": "submitted_for_review"})
	if result.Err() != nil {
		return errors.New("order not found or not awaiting approval by this user")
	}

	_, err := s.orderCollection.UpdateOne(
		ctx,
		bson.M{"_id": orderID},
		bson.M{"$set": bson.M{"status": "approved"}},
	)
	return err
}

func (s *OrderService) ProvideFeedback(orderID, userID primitive.ObjectID, feedback string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify order belongs to the user and is in the correct status
	result := s.orderCollection.FindOne(ctx, bson.M{"_id": orderID, "user_id": userID, "status": "submitted_for_review"})
	if result.Err() != nil {
		return errors.New("order not found or not awaiting feedback from this user")
	}

	_, err := s.orderCollection.UpdateOne(
		ctx,
		bson.M{"_id": orderID},
		bson.M{"$set": bson.M{"status": "feedback", "feedback": feedback}},
	)
	return err
}

func (s *OrderService) WriterAssignmentResponse(orderID, writerID primitive.ObjectID, accept bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if accept {
		// Writer accepts: set status to assigned and update assignment_date
		_, err := s.orderCollection.UpdateOne(
			ctx,
			bson.M{"_id": orderID, "writer_id": writerID, "status": "awaiting_asign_acceptance"},
			bson.M{"$set": bson.M{"status": "assigned", "assignment_date": time.Now()}},
		)
		return err
	} else {
		// Writer declines: set status back to paid, clear writer_id and assignment_date
		_, err := s.orderCollection.UpdateOne(
			ctx,
			bson.M{"_id": orderID, "writer_id": writerID, "status": "awaiting_asign_acceptance"},
			bson.M{"$set": bson.M{"status": "paid"}, "$unset": bson.M{"writer_id": "", "assignment_date": ""}},
		)
		return err
	}
}
