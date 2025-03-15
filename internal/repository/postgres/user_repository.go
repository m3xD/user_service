package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"time"
	"user_service/internal/models"
	"user_service/internal/repository"
	"user_service/internal/util"
)

type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(db *sqlx.DB) repository.UserRepository {
	return &userRepository{db: db}
}

var (
	ErrUserNotFound = errors.New("user not found")
)

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `
        INSERT INTO users (id, email, password, full_name, role, avatar, phone, status, created_at, updated_at)
        VALUES (:id, :email, :password, :full_name, :role, :avatar, :phone, :status, :created_at, :updated_at)
    `
	_, err := r.db.NamedExec(query, user)

	if err != nil {
		return err
	}

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
        SELECT id, email, password, full_name, role, avatar, phone, status, created_at, updated_at
        FROM users WHERE id = $1
    `

	var user models.User

	err := r.db.QueryRowx(query, id).StructScan(&user)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
        SELECT id, email, password, full_name, role, avatar, phone, status, created_at, updated_at
        FROM users WHERE email = $1
    `

	var user models.User
	err := r.db.QueryRowx(query, email).StructScan(&user)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	query := `
        UPDATE users
        SET full_name = :full_name, avatar = :avatar, phone = :phone, status = :status, updated_at = :updated_at, password = :password, role = :role, email = :email, last_login = :last_login, last_activity = :last_activity
        WHERE id = :id
    `

	user.UpdatedAt = time.Now()

	result, err := r.db.NamedExec(query, user)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM users WHERE id = $1"

	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *userRepository) List(ctx context.Context, params util.PaginationParams) ([]*models.User, int64, error) {
	query, countQuery, args, err := util.BuildUserListQuery(params)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build query: %w", err)
	}

	// Get total count
	var total int64
	err = r.db.QueryRow(countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	// Execute the main query
	rows, err := r.db.Queryx(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err = rows.StructScan(&user)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) CountUser(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM users`

	var count int
	err := r.db.Get(&count, query)
	if err != nil {
		return 0, err
	}

	return count, nil
}
