package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type OrderLevel struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name                 string             `bson:"name" json:"name" binding:"required"`
	Description          string             `bson:"description" json:"description"`
	LevelPriceMultiplier float64            `bson:"level_price_multiplier" json:"level_price_multiplier" binding:"required"`
}
