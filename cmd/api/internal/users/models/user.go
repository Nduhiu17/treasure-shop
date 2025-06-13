package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Email    string             `bson:"email" json:"email"`
	Password string             `bson:"password" json:"password"`
	Roles    []string           `bson:"roles" json:"roles"` // e.g., ["user", "writer", "admin", "super_admin"]
	Tier     string             `bson:"tier,omitempty" json:"tier,omitempty"` // For writers: basic, top, premium
	// Add other user-related fields
}