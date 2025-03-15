package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"time"
	"user_service/internal/models"
	"user_service/internal/repository"
)

type authRepository struct {
	db *sqlx.DB
}

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

func (a authRepository) Create(ctx context.Context, token *models.RefreshTokenData) error {
	sql := `INSERT INTO refresh_tokens (id, user_id, token, expires_at, issued_at, is_revoked) VALUES ($1, $2, $3, $4, $5, $6)`

	if token.ID == "" {
		token.ID = uuid.New().String()
	}

	_, err := a.db.ExecContext(ctx, sql, token.ID, token.UserID, token.Token, token.ExpiresAt, token.IssuedAt, token.IsRevoked)

	return err
}

func (a authRepository) GetByToken(ctx context.Context, token string) (*models.RefreshTokenData, error) {
	var refreshToken models.RefreshTokenData
	query := `SELECT id, user_id, token, expires_at, issued_at, is_revoked FROM refresh_tokens WHERE token = $1`
	err := a.db.QueryRowx(query, token).StructScan(&refreshToken)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	if time.Now().After(refreshToken.ExpiresAt) {
		return nil, ErrExpiredToken
	}

	return &refreshToken, nil
}

func (a authRepository) RevokeToken(ctx context.Context, token string) error {
	query := `UPDATE refresh_tokens SET is_revoked = true WHERE token = $1`
	_, err := a.db.ExecContext(ctx, query, token)

	return err
}

func (a authRepository) RevokeAllTokens(ctx context.Context, userID string) error {
	query := `UPDATE refresh_tokens SET is_revoked = true WHERE user_id = $1`
	_, err := a.db.ExecContext(ctx, query, userID)

	return err
}

func (a authRepository) DeleteExpiredTokens(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < $1`
	_, err := a.db.ExecContext(ctx, query, time.Now())

	return err
}

// NewAuthRepository creates a new PostgreSQL auth repository
func NewAuthRepository(db *sqlx.DB) repository.AuthRepository {
	return &authRepository{db: db}
}
