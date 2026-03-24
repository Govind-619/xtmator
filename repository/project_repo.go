package repository

import (
	"database/sql"
	"fmt"

	"github.com/Govind-619/xtmator/domain"
)

// ProjectRepository provides CRUD for user-owned projects.
type ProjectRepository interface {
	Create(userID int64, name, clientName, location string) (*domain.Project, error)
	ListByUser(userID int64) ([]domain.Project, error)
	GetByID(id, userID int64) (*domain.Project, error)
	GetByShareToken(token string) (*domain.Project, error)
	Update(p *domain.Project) error
	Delete(id, userID int64) error
}

type projectRepo struct{ db *sql.DB }

// NewProjectRepository returns a SQLite-backed ProjectRepository.
func NewProjectRepository(db *sql.DB) ProjectRepository { return &projectRepo{db: db} }

func (r *projectRepo) Create(userID int64, name, clientName, location string) (*domain.Project, error) {
	res, err := r.db.Exec(
		`INSERT INTO projects (user_id, name, client_name, location) VALUES (?, ?, ?, ?)`,
		userID, name, clientName, location,
	)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	id, _ := res.LastInsertId()
	return r.GetByID(id, userID)
}

func (r *projectRepo) ListByUser(userID int64) ([]domain.Project, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, name, client_name, location, cost_index, COALESCE(share_token, ''), created_at, updated_at
		 FROM projects WHERE user_id = ? ORDER BY updated_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Project
	for rows.Next() {
		var p domain.Project
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.ClientName, &p.Location, &p.CostIndex, &p.ShareToken, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (r *projectRepo) GetByID(id, userID int64) (*domain.Project, error) {
	p := &domain.Project{}
	err := r.db.QueryRow(
		`SELECT id, user_id, name, client_name, location, cost_index, COALESCE(share_token, ''), created_at, updated_at
		 FROM projects WHERE id = ? AND user_id = ?`, id, userID,
	).Scan(&p.ID, &p.UserID, &p.Name, &p.ClientName, &p.Location, &p.CostIndex, &p.ShareToken, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func (r *projectRepo) GetByShareToken(token string) (*domain.Project, error) {
	p := &domain.Project{}
	err := r.db.QueryRow(
		`SELECT id, user_id, name, client_name, location, cost_index, COALESCE(share_token, ''), created_at, updated_at
		 FROM projects WHERE share_token = ?`, token,
	).Scan(&p.ID, &p.UserID, &p.Name, &p.ClientName, &p.Location, &p.CostIndex, &p.ShareToken, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func (r *projectRepo) Update(p *domain.Project) error {
	_, err := r.db.Exec(
		`UPDATE projects SET name=?, client_name=?, location=?, cost_index=?, share_token=?, updated_at=CURRENT_TIMESTAMP
		 WHERE id=? AND user_id=?`,
		p.Name, p.ClientName, p.Location, p.CostIndex, p.ShareToken, p.ID, p.UserID,
	)
	return err
}

func (r *projectRepo) Delete(id, userID int64) error {
	_, err := r.db.Exec(`DELETE FROM projects WHERE id=? AND user_id=?`, id, userID)
	return err
}
