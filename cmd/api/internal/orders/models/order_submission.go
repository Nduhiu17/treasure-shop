package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderSubmission struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrderID         primitive.ObjectID `bson:"order_id" json:"order_id"`
	SubmissionDate  time.Time          `bson:"submission_date" json:"submission_date"`
	Description     string             `bson:"description" json:"description"`
	SubmissionFile  string             `bson:"submission_file" json:"submission_file"`
	SubmissionTrial int                `bson:"submission_trial" json:"submission_trial"`
}
