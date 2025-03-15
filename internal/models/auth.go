package models

import (
	"context"
	"errors"
	"time"
)

type AuthService interface {
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	Logout(ctx context.Context, token string) error
	RefreshToken(ctx context.Context, token string) (string, string, error)
	SaveToken(ctx context.Context, token string, userID string) error
	LogoutAll(ctx context.Context, userID string) error
}

// LoginRequest represents the login credentials
type LoginRequest struct {
	Email string `json:"email"`

	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required"`
	Phone    string `json:"phone" validate:"required"`
}

var (
	ErrEmailEmpty    = errors.New("email is required")
	ErrPasswordEmpty = errors.New("password is required")
)

// validate login request here
func (r LoginRequest) Validate() error {
	if r.Email == "" {
		return ErrEmailEmpty
	}

	if r.Password == "" {
		return ErrPasswordEmpty
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

type RefreshTokenData struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires"`
	IssuedAt  time.Time `json:"issued"`
	IsRevoked bool      `json:"is_revoked"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type AuthenticationError struct {
	Status string `json:"status,omitempty"`

	Timestamp time.Time `json:"timestamp,omitempty"`

	Message string `json:"message,omitempty"`

	Path string `json:"path,omitempty"`
}

// LoginResponse represents the response for successful login
type LoginResponse struct {
	Token string `json:"token,omitempty"`

	RefreshToken string `json:"refreshToken,omitempty"`

	User *UserSummary `json:"user,omitempty"`
}
