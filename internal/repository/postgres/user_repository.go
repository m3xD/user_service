package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"time"
	"user_service/internal/models"
	"user_service/internal/repository"
)

type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(db *sqlx.DB) repository.UserRepository {
	return &userRepository{db: db}
}

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
		return nil, errors.New("user not found")
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
		return nil, errors.New("user not found")
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	query := `
        UPDATE users
        SET full_name = :full_name, avatar = :avatar, phone = :phone, status = :status, updated_at = :updated_at
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
		return errors.New("user not found")
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
		return errors.New("user not found")
	}

	return nil
}

func (r *userRepository) List(ctx context.Context, page, pageSize int) ([]*models.User, error) {
	query := `
        SELECT id, email, password, full_name, role, avatar, phone, status, created_at, updated_at
        FROM users
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `

	rows, err := r.db.Queryx(query, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err = rows.StructScan(&user)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
