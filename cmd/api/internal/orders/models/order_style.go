package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type OrderStyle struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name" json:"name" binding:"required"`
	Description string             `bson:"description" json:"description"`
}
