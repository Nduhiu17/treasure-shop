package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderFeedback struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrderID      primitive.ObjectID `bson:"order_id" json:"order_id"`
	Feedback     string             `bson:"feedback" json:"feedback"`
	FeedbackFile string             `bson:"feedback_file" json:"feedback_file"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}
