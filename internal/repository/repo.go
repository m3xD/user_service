package repository

import (
	"context"
	"user_service/internal/models"
	"user_service/internal/util"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params util.PaginationParams) ([]*models.User, int64, error)
	CountUser(ctx context.Context) (int, error)
}

type AuthRepository interface {
	Create(ctx context.Context, token *models.RefreshTokenData) error
	GetByToken(ctx context.Context, token string) (*models.RefreshTokenData, error)
	RevokeToken(ctx context.Context, token string) error
	RevokeAllTokens(ctx context.Context, userID string) error
	DeleteExpiredTokens(ctx context.Context) error
}
