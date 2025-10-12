package user

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	Create(ctx context.Context, u *User) (int64, error)
	IsEmailUnique(ctx context.Context, u *User) (bool, error)
	//GetByID(ctx context.Context, id int64) (*User, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, u *User) (int64, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("hash password error: %w", err)
	}
	isUnique, err := r.IsEmailUnique(ctx, u)
	if err != nil {
		return 0, fmt.Errorf("email uniqueness error: %w", err)
	}
	if !isUnique {
		return 0, fmt.Errorf("email already exists")
	}
	const q = `INSERT INTO users (email, name, password_hash) VALUES ($1,$2,$3) RETURNING id`
	var id int64
	if err := r.db.QueryRowxContext(ctx, q, u.Email, u.Name, hashedPassword).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *repository) IsEmailUnique(ctx context.Context, u *User) (bool, error) {
	var exist bool
	const q = "SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)"
	if err := r.db.QueryRowContext(ctx, q, u.Email).Scan(&exist); err != nil {
		return false, fmt.Errorf("checking email uniaueness: %w", err)
	}
	return !exist, nil

}
