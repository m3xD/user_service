package models

import (
	"errors"
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
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	Password     string    `json:"-" db:"password"` // Not exposed in JSON
	FullName     string    `json:"fullName" db:"full_name"`
	Role         string    `json:"role" db:"role"` // user, admin
	Avatar       string    `json:"avatar,omitempty" db:"avatar"`
	Phone        string    `json:"phone" db:"phone"`
	Status       string    `json:"status" db:"status"` // active, inactive
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
	LastLogin    time.Time `json:"lastLogin,omitempty"`
	LastActivity time.Time `json:"lastActivity,omitempty"`
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

type UserSummary struct {
	Id string `json:"id,omitempty"`

	Name string `json:"name,omitempty"`

	Email string `json:"email,omitempty"`

	Role string `json:"role,omitempty"`
}

type UserPage struct {
	Content []*User `json:"content,omitempty"`

	Pageable *PageableObject `json:"pageable,omitempty"`

	TotalPages int32 `json:"totalPages,omitempty"`

	TotalElements int32 `json:"totalElements,omitempty"`

	Last bool `json:"last,omitempty"`

	First bool `json:"first,omitempty"`

	Sort *SortObject `json:"sort,omitempty"`

	Number int32 `json:"number,omitempty"`

	NumberOfElements int32 `json:"numberOfElements,omitempty"`

	Size int32 `json:"size,omitempty"`

	Empty bool `json:"empty,omitempty"`

	ActiveUsers int32 `json:"activeUsers,omitempty"`

	InactiveUsers int32 `json:"inactiveUsers,omitempty"`
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8"`
}

func (c *ChangePasswordInput) Validate() error {

	if c.CurrentPassword == "" || c.NewPassword == "" {
		return errors.New("invalid request")
	}

	return nil
}
