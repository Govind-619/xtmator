package repository

import (
	"database/sql"
	"fmt"

	"github.com/Govind-619/xtmator/domain"
)

// UserRepository provides persistence for user accounts.
type UserRepository interface {
	Create(name, email, passwordHash string) (*domain.User, error)
	FindByEmail(email string) (*domain.User, error)
	FindByID(id int64) (*domain.User, error)
}

type userRepo struct{ db *sql.DB }

// NewUserRepository returns a SQLite-backed UserRepository.
func NewUserRepository(db *sql.DB) UserRepository { return &userRepo{db: db} }

func (r *userRepo) Create(name, email, passwordHash string) (*domain.User, error) {
	res, err := r.db.Exec(
		`INSERT INTO users (name, email, password_hash) VALUES (?, ?, ?)`,
		name, email, passwordHash,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	id, _ := res.LastInsertId()
	return r.FindByID(id)
}

func (r *userRepo) FindByEmail(email string) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow(
		`SELECT id, name, email, password_hash, created_at FROM users WHERE email = ?`, email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *userRepo) FindByID(id int64) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow(
		`SELECT id, name, email, password_hash, created_at FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}
