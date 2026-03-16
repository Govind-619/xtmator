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
	FindByGoogleID(googleID string) (*domain.User, error)
	FindOrCreateGoogleUser(name, email, googleID string) (*domain.User, error)
}

type userRepo struct{ db *sql.DB }

func NewUserRepository(db *sql.DB) UserRepository { return &userRepo{db: db} }

func (r *userRepo) scan(row *sql.Row) (*domain.User, error) {
	u := &domain.User{}
	var hash, googleID, provider sql.NullString
	err := row.Scan(&u.ID, &u.Name, &u.Email, &hash, &googleID, &provider, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u.PasswordHash = hash.String
	u.GoogleID = googleID.String
	u.AuthProvider = provider.String
	return u, nil
}

func (r *userRepo) Create(name, email, passwordHash string) (*domain.User, error) {
	res, err := r.db.Exec(
		`INSERT INTO users (name, email, password_hash, auth_provider) VALUES (?, ?, ?, 'email')`,
		name, email, passwordHash,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	id, _ := res.LastInsertId()
	return r.FindByID(id)
}

func (r *userRepo) FindByEmail(email string) (*domain.User, error) {
	return r.scan(r.db.QueryRow(
		`SELECT id, name, email, password_hash, google_id, auth_provider, created_at FROM users WHERE email = ?`, email,
	))
}

func (r *userRepo) FindByID(id int64) (*domain.User, error) {
	return r.scan(r.db.QueryRow(
		`SELECT id, name, email, password_hash, google_id, auth_provider, created_at FROM users WHERE id = ?`, id,
	))
}

func (r *userRepo) FindByGoogleID(googleID string) (*domain.User, error) {
	return r.scan(r.db.QueryRow(
		`SELECT id, name, email, password_hash, google_id, auth_provider, created_at FROM users WHERE google_id = ?`, googleID,
	))
}

// FindOrCreateGoogleUser looks up by google_id or email, creating the account if needed.
func (r *userRepo) FindOrCreateGoogleUser(name, email, googleID string) (*domain.User, error) {
	// 1. Try by Google ID
	if u, err := r.FindByGoogleID(googleID); u != nil || err != nil {
		return u, err
	}
	// 2. Try by email (account exists, link Google ID to it)
	existing, err := r.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		_, err = r.db.Exec(
			`UPDATE users SET google_id=?, auth_provider='google' WHERE id=?`,
			googleID, existing.ID,
		)
		if err != nil {
			return nil, err
		}
		return r.FindByID(existing.ID)
	}
	// 3. New user — create with Google auth
	res, err := r.db.Exec(
		`INSERT INTO users (name, email, google_id, auth_provider) VALUES (?, ?, ?, 'google')`,
		name, email, googleID,
	)
	if err != nil {
		return nil, fmt.Errorf("create google user: %w", err)
	}
	id, _ := res.LastInsertId()
	return r.FindByID(id)
}
