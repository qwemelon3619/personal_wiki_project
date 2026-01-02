package model

import "time"

// User represents a user in the photo gallery system
type User struct {
	UserID       string    `json:"userId" bson:"_id, omitempty"`
	Email        string    `json:"email" bson:"email"`     // Unique email
	Name         string    `json:"name" bson:"name"`       // Display name
	PasswordHash string    `json:"-" bson:"password_hash"` // bcrypt hash (if used)
	Roles        []string  `json:"roles" bson:"roles"`     // e.g., ["user","admin"]
	IsActive     bool      `json:"isActive" bson:"is_active"`
	LastSeen     time.Time `json:"lastSeen" bson:"last_seen"`
	UpdatedAt    time.Time `json:"updatedAt" bson:"updated_at"`
	CreatedAt    time.Time `json:"createdAt" bson:"created_at"`
}
