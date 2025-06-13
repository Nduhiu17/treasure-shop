package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Email     string             `bson:"email" json:"email" binding:"required"`
	Username  string             `bson:"username" json:"username" binding:"required"`
	FirstName string             `bson:"first_name" json:"first_name" binding:"required"`
	LastName  string             `bson:"last_name" json:"last_name" binding:"required"`
	Password  string             `bson:"password" json:"-" binding:"required"`
	Tier      string             `bson:"tier,omitempty" json:"tier,omitempty"`
	Roles     []string           `bson:"-" json:"roles,omitempty"`
}

// Role struct for roles collection
// Each role has an ID and a role name
// e.g., {"_id": ObjectId, "name": "admin"}
type Role struct {
	ID   primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name string             `bson:"name" json:"name" binding:"required"`
}

// UserRole struct for user_roles collection
// Each user_role links a user to a role
// e.g., {"_id": ObjectId, "user_id": ObjectId, "role_id": ObjectId}
type UserRole struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID primitive.ObjectID `bson:"user_id" json:"user_id" binding:"required"`
	RoleID primitive.ObjectID `bson:"role_id" json:"role_id" binding:"required"`
}
