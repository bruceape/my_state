package models

import (
	"database/sql"
	_ "embed"
	"errors"
	"time"
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

func (r *UserRepository) CreateUser(user *User) error {
	err := r.db.QueryRow(createUserQuery, user.Email, user.PasswordHash).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	return err
}

//go:embed find_user_by_email.sql
var findUserByEmailQuery string

func (r *UserRepository) FindUserByEmail(email string) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(findUserByEmailQuery, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}

	return user, err
}
