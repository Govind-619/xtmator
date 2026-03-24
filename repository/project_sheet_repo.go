package repository

import (
	"database/sql"
	"fmt"

	"github.com/Govind-619/xtmator/domain"
)

type ProjectSheetRepository interface {
	Create(projectID int64, name string) (*domain.ProjectSheet, error)
	ListByProject(projectID int64) ([]domain.ProjectSheet, error)
	Update(sheet *domain.ProjectSheet) error
	Delete(sheetID, projectID int64) error
	GetByID(sheetID, projectID int64) (*domain.ProjectSheet, error)
}

type projectSheetRepo struct{ db *sql.DB }

func NewProjectSheetRepository(db *sql.DB) ProjectSheetRepository {
	return &projectSheetRepo{db: db}
}

func (r *projectSheetRepo) Create(projectID int64, name string) (*domain.ProjectSheet, error) {
	res, err := r.db.Exec(`INSERT INTO project_sheets (project_id, name) VALUES (?, ?)`, projectID, name)
	if err != nil {
		return nil, fmt.Errorf("create project sheet: %w", err)
	}
	id, _ := res.LastInsertId()
	return r.GetByID(id, projectID)
}

func (r *projectSheetRepo) ListByProject(projectID int64) ([]domain.ProjectSheet, error) {
	rows, err := r.db.Query(
		`SELECT id, project_id, name, created_at, updated_at 
		 FROM project_sheets WHERE project_id = ? ORDER BY id ASC`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sheets []domain.ProjectSheet
	for rows.Next() {
		var s domain.ProjectSheet
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.Name, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		sheets = append(sheets, s)
	}
	return sheets, nil
}

func (r *projectSheetRepo) GetByID(sheetID, projectID int64) (*domain.ProjectSheet, error) {
	var s domain.ProjectSheet
	err := r.db.QueryRow(
		`SELECT id, project_id, name, created_at, updated_at 
		 FROM project_sheets WHERE id = ? AND project_id = ?`,
		sheetID, projectID,
	).Scan(&s.ID, &s.ProjectID, &s.Name, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *projectSheetRepo) Update(sheet *domain.ProjectSheet) error {
	_, err := r.db.Exec(
		`UPDATE project_sheets SET name=?, updated_at=CURRENT_TIMESTAMP WHERE id=? AND project_id=?`,
		sheet.Name, sheet.ID, sheet.ProjectID,
	)
	return err
}

func (r *projectSheetRepo) Delete(sheetID, projectID int64) error {
	_, err := r.db.Exec(`DELETE FROM project_sheets WHERE id=? AND project_id=?`, sheetID, projectID)
	return err
}
