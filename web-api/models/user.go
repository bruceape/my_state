package models

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db}
}

//go:embed create_user.sql
var createUserQuery string

func (r *UserRepository) CreateUser(ctx context.Context, user *User) error {
	err := r.db.QueryRowContext(ctx, createUserQuery, user.Email, user.PasswordHash).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return ErrUserExists
		}
	}

	return err
}

//go:embed find_user_by_email.sql
var findUserByEmailQuery string

func (r *UserRepository) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	user := &User{}
	err := r.db.QueryRowContext(ctx, findUserByEmailQuery, email).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err == nil {
		return user, nil
	}

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

	return nil, err
}
