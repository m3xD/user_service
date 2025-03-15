package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"time"
	"user_service/internal/models"
	"user_service/internal/repository"
	"user_service/internal/repository/postgres"
	"user_service/internal/util"
)

type authService struct {
	userRepo   repository.UserRepository
	authRepo   repository.AuthRepository
	jwtService *util.JwtImpl
	log        *zap.Logger
}

func NewAuthService(userRepo repository.UserRepository, jwtService *util.JwtImpl, log *zap.Logger, authRepo repository.AuthRepository) *authService {
	return &authService{userRepo: userRepo, jwtService: jwtService, log: log, authRepo: authRepo}
}

var (
	ErrExpiredToken = errors.New("expired token")
)

func (s *authService) Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
	// Validate credentials with user service
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, postgres.ErrUserNotFound) {
			s.log.Error("[AuthService][Login] user not found", zap.Error(err))
			return nil, ErrUserNotFound
		}
		s.log.Error("[AuthService][Login] failed to validate credentials", zap.Error(err))
		return nil, ErrInvalidEmailOrPassword
	}

	// Compare password

	// Generate tokens
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		s.log.Error("[AuthService][Login] failed to generate access token", zap.Error(err))
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		s.log.Error("[AuthService][Login] failed to generate refresh token", zap.Error(err))
		return nil, err
	}

	// Store refresh token
	err = s.authRepo.Create(ctx, &models.RefreshTokenData{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now(),
		IssuedAt:  time.Now().Add(24 * time.Hour),
		IsRevoked: false,
	})
	if err != nil {
		s.log.Error("[AuthService][Login] failed to store refresh token", zap.Error(err))
		return nil, err
	}

	// Create login response
	loginResponse := &models.LoginResponse{
		RefreshToken: refreshToken,
		Token:        accessToken,
		User: &models.UserSummary{
			Id:    user.ID,
			Name:  user.FullName,
			Email: user.Email,
			Role:  user.Role,
		},
	}

	return loginResponse, nil
}

func (s *authService) RefreshToken(ctx context.Context, token string) (string, string, error) {
	claims, err := s.jwtService.ValidateRefreshToken(token)
	if err != nil {
		s.log.Error("[AuthService][RefreshToken] failed to validate refresh token", zap.Error(err))
		return "", "", err
	}

	storedToken, err := s.authRepo.GetByToken(ctx, token)
	if err != nil {
		s.log.Error("[AuthService][RefreshToken] failed to get refresh token", zap.Error(err))
		return "", "", err
	}

	if storedToken.IsRevoked {
		return "", "", ErrExpiredToken
	}

	if err := s.authRepo.RevokeToken(ctx, token); err != nil {
		s.log.Error("[AuthService][RefreshToken] failed to revoke token", zap.Error(err))
		return "", "", err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(claims["userID"].(string), claims["role"].(string))
	if err != nil {
		s.log.Error("[AuthService][RefreshToken] failed to generate refresh token", zap.Error(err))
		return "", "", err
	}

	err = s.authRepo.Create(ctx, &models.RefreshTokenData{
		ID:        uuid.New().String(),
		UserID:    claims["userID"].(string),
		Token:     refreshToken,
		ExpiresAt: time.Now(),
		IssuedAt:  time.Now().Add(24 * time.Hour),
		IsRevoked: false,
	})

	if err != nil {
		s.log.Error("[AuthService][RefreshToken] failed to store refresh token", zap.Error(err))
		return "", "", err
	}

	// Generate new access token
	accessToken, err := s.jwtService.GenerateAccessToken(claims["userID"].(string), claims["role"].(string))
	if err != nil {
		s.log.Error("[AuthService][RefreshToken] failed to generate access token", zap.Error(err))
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *authService) Logout(ctx context.Context, token string) error {
	err := s.authRepo.RevokeToken(ctx, token)
	if err != nil {
		s.log.Error("[AuthService][Logout] failed to revoke token", zap.Error(err))
		return err
	}

	return nil
}

func (s *authService) SaveToken(ctx context.Context, token string, userID string) error {
	err := s.authRepo.Create(ctx, &models.RefreshTokenData{
		ID:        uuid.New().String(),
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IssuedAt:  time.Now(),
		UserID:    userID,
		IsRevoked: false,
	})

	if err != nil {
		s.log.Error("[AuthService][SaveToken] failed to save token", zap.Error(err))
		return err
	}

	return nil
}

func (s *authService) LogoutAll(ctx context.Context, userID string) error {
	err := s.authRepo.RevokeAllTokens(ctx, userID)
	if err != nil {
		s.log.Error("[AuthService][LogoutAll] failed to revoke all tokens", zap.Error(err))
		return err
	}
	return nil
}
