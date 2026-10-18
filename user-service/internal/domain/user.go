package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name            string             `json:"name" bson:"name"`
	Email           string             `json:"email" bson:"email"`
	PasswordHash    string             `json:"password_hash" bson:"password_hash"`
	Role            UserRole           `json:"role" bson:"role"`
	Status          UserStatus         `json:"status" bson:"status"`
	Version         int64              `json:"version" bson:"version"`
	EmailVerifiedAt *time.Time         `json:"emailVerifiedAt,omitempty" bson:"emailVerifiedAt,omitempty"`
	CreatedAt       time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt       time.Time          `json:"updatedAt" bson:"updatedAt"`
}
