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
	"go.mongodb.org/mongo-driver/mongo/options"
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
	order.ApplyFeedbackRequests = 0 // Default to zero on creation
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

func (s *OrderService) GetAllOrdersPaginated(page, pageSize int) ([]models.Order, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	total, err := s.orderCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.M{"created_at": -1})
	cursor, err := s.orderCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, 0, err
	}
	return orders, total, nil
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

// GetOrdersFiltered returns orders filtered by user_id, writer_id, and/or status (all are optional)
func (s *OrderService) GetOrdersFiltered(userID, writerID *primitive.ObjectID, status *string, page, pageSize int) ([]models.Order, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	if userID != nil {
		filter["user_id"] = *userID
	}
	if writerID != nil {
		filter["writer_id"] = *writerID
	}
	if status != nil && *status != "" {
		filter["status"] = *status
	}

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)
	total, err := s.orderCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	findOpts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.M{"created_at": -1})
	cursor, err := s.orderCollection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, 0, err
	}
	return orders, total, nil
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

	// Allow reassignment if order is in 'paid' or 'feedback' status
	filter := bson.M{"_id": orderID, "$or": []bson.M{{"status": "paid"}, {"status": "feedback"}}}
	update := bson.M{"$set": bson.M{"writer_id": writerID, "status": "awaiting_asign_acceptance", "assignment_date": time.Now()}}
	_, err = s.orderCollection.UpdateOne(ctx, filter, update)
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
		bson.M{"$set": bson.M{"status": "submitted_for_review", "content": content, "submission_date": time.Now()}},
	)
	if err != nil {
		return err
	}

	// In a real application, you would likely create a 'Review' record here as well.
	// This could include the content submitted by the writer, the order ID, and any other relevant details.

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
		bson.M{"$set": bson.M{"status": "approved", "approval_date": time.Now()}},
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

	// Increment apply_feedback_requests when feedback is requested
	// Take the current number of requests of the order and increment it by 1
	feedbackCount := 0
	err := s.orderCollection.FindOne(ctx, bson.M{"_id": orderID}).Decode(&bson.M{"apply_feedback_requests": &feedbackCount})
	if err != nil {
		return errors.New("failed to retrieve current feedback request count")
	}

	if feedbackCount >= 4 {
		return errors.New("feedback request limit reached for this order")
	}
	// Update the order status to 'feedback' and set the feedback
	_, err = s.orderCollection.UpdateOne(
		ctx,
		bson.M{"_id": orderID, "user_id": userID, "status": "submitted_for_review"},
		bson.M{"$set": bson.M{"status": "feedback", "feedback": feedback, "feedback_date": time.Now()}, "$inc": bson.M{"apply_feedback_requests": feedbackCount + 1}},
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
			bson.M{"$set": bson.M{"status": "assigned", "assignment_acceptance_date": time.Now()}},
		)
		return err
	} else {
		// Writer declines: set status back to paid, clear writer_id and assignment_date
		_, err := s.orderCollection.UpdateOne(
			ctx,
			bson.M{"_id": orderID, "writer_id": writerID, "status": "awaiting_asign_acceptance"},
			bson.M{"$set": bson.M{"status": "paid"}, "$unset": bson.M{"writer_id": "", "assignment_date": "", "assignment_decline_date": time.Now()}},
		)
		return err
	}
}

// Helper: populate LevelName for orders
func PopulateOrderLevelNames(orders []models.Order, orderLevelService *OrderLevelService) []models.Order {
	for i, order := range orders {
		if order.OrderLevelID.IsZero() {
			orders[i].LevelName = ""
			continue
		}
		level, err := orderLevelService.GetByID(order.OrderLevelID)
		if err == nil && level != nil {
			orders[i].LevelName = level.Name
		} else {
			orders[i].LevelName = ""
		}
	}
	return orders
}

// Helper: populate OrderPagesName for orders
func PopulateOrderPagesNames(orders []models.Order, orderPagesService *OrderPagesService) []models.Order {
	for i, order := range orders {
		if order.OrderPagesID.IsZero() {
			orders[i].OrderPagesName = ""
			continue
		}
		pages, err := orderPagesService.GetByID(order.OrderPagesID)
		if err == nil && pages != nil {
			orders[i].OrderPagesName = pages.Name
		} else {
			orders[i].OrderPagesName = ""
		}
	}
	return orders
}
