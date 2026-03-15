package repository

import (
	"database/sql"

	"github.com/Govind-619/xtmator/domain"
)

// DSRRepository provides read access to the DSR items catalogue.
type DSRRepository interface {
	ListCategories() ([]string, error)
	ListByCategory(category string) ([]domain.DSRItem, error)
	GetByID(id int64) (*domain.DSRItem, error)
}

type dsrRepo struct{ db *sql.DB }

// NewDSRRepository returns a SQLite-backed DSRRepository.
func NewDSRRepository(db *sql.DB) DSRRepository { return &dsrRepo{db: db} }

func (r *dsrRepo) ListCategories() ([]string, error) {
	rows, err := r.db.Query(`SELECT DISTINCT category FROM dsr_items ORDER BY category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []string
	for rows.Next() {
		var c string
		rows.Scan(&c)
		cats = append(cats, c)
	}
	return cats, nil
}

func (r *dsrRepo) ListByCategory(category string) ([]domain.DSRItem, error) {
	rows, err := r.db.Query(
		`SELECT id, category, code, description, unit, rate
		 FROM dsr_items WHERE category = ? ORDER BY code`, category,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.DSRItem
	for rows.Next() {
		var item domain.DSRItem
		rows.Scan(&item.ID, &item.Category, &item.Code, &item.Description, &item.Unit, &item.Rate)
		items = append(items, item)
	}
	return items, nil
}

func (r *dsrRepo) GetByID(id int64) (*domain.DSRItem, error) {
	item := &domain.DSRItem{}
	err := r.db.QueryRow(
		`SELECT id, category, code, description, unit, rate FROM dsr_items WHERE id = ?`, id,
	).Scan(&item.ID, &item.Category, &item.Code, &item.Description, &item.Unit, &item.Rate)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return item, err
}
