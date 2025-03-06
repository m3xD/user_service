package service

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"user_service/internal/models"
	"user_service/internal/repository"
	"user_service/internal/util"
)

type authService struct {
	userRepo   repository.UserRepository
	jwtService *util.JwtImpl
	log        *zap.Logger
}

func NewAuthService(userRepo repository.UserRepository, jwtService *util.JwtImpl, log *zap.Logger) *authService {
	return &authService{userRepo: userRepo, jwtService: jwtService, log: log}
}

func (s *authService) Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
	// Validate credentials with user service
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			s.log.Error("[AuthService][Login] user not found", zap.Error(err))
			return nil, errors.New("invalid credentials")
		}
		s.log.Error("[AuthService][Login] failed to validate credentials", zap.Error(err))
		return nil, errors.New("invalid credentials")
	}

	// Compare password

	// Generate tokens
	accessToken, err := s.jwtService.GenerateToken(user.ID, user.Role)
	if err != nil {
		s.log.Error("[AuthService][Login] failed to generate access token", zap.Error(err))
		return nil, err
	}

	// Create login response
	loginResponse := &models.LoginResponse{
		Token: accessToken,
		User:  user,
	}

	return loginResponse, nil
}
