package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"time"
	"user_service/internal/models"
	"user_service/internal/repository"
)

type UserService interface {
	Create(ctx context.Context, input models.CreateUserInput) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
	Update(ctx context.Context, id string, input models.UpdateUserInput) (*models.User, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page, pageSize int) ([]*models.User, error)
	Validate(ctx context.Context, email, password string) (*models.User, error)
}

type userService struct {
	repo repository.UserRepository
	log  *zap.Logger
}

var (
	ErrUserNotFound = errors.New("user not found")
	ErrorUserExists = errors.New("user already exists")
	ErrorGetUser    = errors.New("failed to get user")
	ErrorHashing    = errors.New("failed to hash password")
	ErrorCreating   = errors.New("failed to create user")
	ErrorUpdating   = errors.New("failed to update user")
	ErrorDeleting   = errors.New("failed to delete user")
	ErrorListing    = errors.New("failed to list users")
)

func (s userService) Create(ctx context.Context, input models.CreateUserInput) (*models.User, error) {
	existing, err := s.repo.GetByEmail(ctx, input.Email)
	if err == nil && existing != nil {
		s.log.Error("[Service][Create] user already exists", zap.Error(err))
		return nil, ErrorUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("[Service][Create] failed to hash password", zap.Error(err))
		return nil, ErrorHashing
	}

	user := &models.User{
		ID:        uuid.New().String(),
		Email:     input.Email,
		Password:  string(hashedPassword),
		FullName:  input.FullName,
		Role:      input.Role,
		Phone:     input.Phone,
		Avatar:    "default.jpg", // temporary
		Status:    models.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		s.log.Error("[Service][Create] failed to create user", zap.Error(err))
		return nil, ErrorCreating
	}

	return user, nil
}

func (s userService) GetByID(ctx context.Context, id string) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("[Service][GetByID] failed to get user", zap.Error(err))
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s userService) Update(ctx context.Context, id string, input models.UpdateUserInput) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("[Service][Update] failed to get user", zap.Error(err))
		return nil, ErrorGetUser
	}

	if input.FullName != "" {
		user.FullName = input.FullName
	}
	if input.Phone != "" {
		user.Phone = input.Phone
	}
	if input.Avatar != "" {
		user.Avatar = input.Avatar
	}
	if input.Status != "" {
		user.Status = input.Status
	}

	if err := s.repo.Update(ctx, user); err != nil {
		s.log.Error("[Service][Update] failed to update user", zap.Error(err))
		return nil, ErrorUpdating
	}

	return user, nil
}

func (s userService) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Error("[Service][Delete] failed to delete user", zap.Error(err))
		return ErrorDeleting
	}

	return nil
}

func (s userService) List(ctx context.Context, page, pageSize int) ([]*models.User, error) {
	users, err := s.repo.List(ctx, page, pageSize)
	if err != nil {
		s.log.Error("[Service][List] failed to list users", zap.Error(err))
		return nil, ErrorListing
	}

	return users, nil
}

func (s userService) Validate(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		s.log.Error("[Service][Validate] failed to get user", zap.Error(err))
		return nil, ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		s.log.Error("[Service][Validate] invalid password", zap.Error(err))
		return nil, ErrUserNotFound
	}

	return user, nil
}

func NewUserService(repo repository.UserRepository, log *zap.Logger) UserService {
	return &userService{repo: repo, log: log}
}
