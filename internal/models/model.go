package models

import (
	"time"
)

// Role constants
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// Status constants
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
)

// User represents the user entity
type User struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"` // Not exposed in JSON
	FullName  string    `json:"full_name" db:"full_name"`
	Role      string    `json:"role" db:"role"` // user, admin
	Avatar    string    `json:"avatar,omitempty" db:"avatar"`
	Phone     string    `json:"phone" db:"phone"`
	Status    string    `json:"status" db:"status"` // active, inactive
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateUserInput represents the input for user creation
type CreateUserInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required"`
	Phone    string `json:"phone" validate:"required"`
	Role     string `json:"role" validate:"required,oneof=user admin"`
}

// UpdateUserInput represents the input for user update
type UpdateUserInput struct {
	FullName string `json:"full_name,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	Status   string `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
}

// LoginInput represents the input for user login
type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
