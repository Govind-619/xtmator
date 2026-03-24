package repository

import (
	"database/sql"
	"fmt"

	"github.com/Govind-619/xtmator/domain"
)

type CustomItemRepository interface {
	Create(item *domain.CustomItem) (*domain.CustomItem, error)
	ListByUser(userID int64) ([]domain.CustomItem, error)
	Delete(id, userID int64) error
}

type customItemRepo struct{ db *sql.DB }

func NewCustomItemRepository(db *sql.DB) CustomItemRepository {
	return &customItemRepo{db: db}
}

func (r *customItemRepo) Create(item *domain.CustomItem) (*domain.CustomItem, error) {
	res, err := r.db.Exec(
		`INSERT INTO custom_items (user_id, category, description, unit, rate)
		 VALUES (?, ?, ?, ?, ?)`,
		item.UserID, item.Category, item.Description, item.Unit, item.Rate,
	)
	if err != nil {
		return nil, fmt.Errorf("create custom item: %w", err)
	}
	id, _ := res.LastInsertId()
	item.ID = id
	return item, nil
}

func (r *customItemRepo) ListByUser(userID int64) ([]domain.CustomItem, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, category, description, unit, rate, created_at
		 FROM custom_items WHERE user_id = ? ORDER BY category, description`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.CustomItem
	for rows.Next() {
		var i domain.CustomItem
		if err := rows.Scan(&i.ID, &i.UserID, &i.Category, &i.Description, &i.Unit, &i.Rate, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}

func (r *customItemRepo) Delete(id, userID int64) error {
	_, err := r.db.Exec(`DELETE FROM custom_items WHERE id=? AND user_id=?`, id, userID)
	return err
}
