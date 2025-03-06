package models

import (
	"context"
	"errors"
)

type AuthService interface {
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
}

// LoginRequest represents the login credentials
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required"`
	Phone    string `json:"phone" validate:"required"`
}

// LoginResponse represents the response for successful login
type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// validate login request here
func (r LoginRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}

	if r.Password == "" {
		return errors.New("password is required")
	}

	return nil
}

func (r RegisterRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}

	if r.Password == "" {
		return errors.New("password is required")
	}

	if r.FullName == "" {
		return errors.New("full name is required")
	}

	if r.Phone == "" {
		return errors.New("phone is required")
	}

	return nil
}
