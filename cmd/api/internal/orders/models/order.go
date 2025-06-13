package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID         primitive.ObjectID `bson:"user_id" json:"user_id"`
	OrderTypeID    primitive.ObjectID `bson:"order_type_id" json:"order_type_id"` // Foreign key to OrderType
	Title          string             `bson:"title" json:"title"`
	Description    string             `bson:"description" json:"description"`
	Price          float64            `bson:"price" json:"price"`
	Status         string             `bson:"status" json:"status"` // pending_payment, awaiting_assignment, assigned, in_progress, submitted_for_review, approved, feedback, completed
	WriterID       primitive.ObjectID `bson:"writer_id,omitempty" json:"writer_id,omitempty"`
	SubmissionDate *time.Time         `bson:"submission_date,omitempty" json:"submission_date,omitempty"`
	Feedback       string             `bson:"feedback,omitempty" json:"feedback,omitempty"`
	CreatedAt      time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt      time.Time          `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}
