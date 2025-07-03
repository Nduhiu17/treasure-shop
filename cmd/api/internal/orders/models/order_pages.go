package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type OrderPages struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name          string             `bson:"name" json:"name" binding:"required"`
	Description   string             `bson:"description" json:"description"`
	NumberOfPages int                `bson:"number_of_pages" json:"number_of_pages" binding:"required"`
}
