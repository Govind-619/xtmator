package repository

import (
	"database/sql"
	"fmt"

	"github.com/Govind-619/xtmator/domain"
)

// BOQRepository manages Bill of Quantities entries within a project.
type BOQRepository interface {
	AddEntry(entry *domain.BOQEntry) (*domain.BOQEntry, error)
	ListByProject(projectID, sheetID int64) ([]domain.BOQEntry, error)
	GetEntry(id, projectID int64) (*domain.BOQEntry, error)
	UpdateEntry(entry *domain.BOQEntry) error
	DeleteEntry(id, projectID int64) error
	NextItemNo(projectID, sheetID int64) int
}

type boqRepo struct{ db *sql.DB }

// NewBOQRepository returns a SQLite-backed BOQRepository.
func NewBOQRepository(db *sql.DB) BOQRepository { return &boqRepo{db: db} }

func (r *boqRepo) NextItemNo(projectID, sheetID int64) int {
	var maxNo int
	r.db.QueryRow(
		`SELECT COALESCE(MAX(item_no), 0) FROM boq_entries WHERE project_id = ? AND sheet_id = ?`, projectID, sheetID,
	).Scan(&maxNo)
	return maxNo + 1
}

func (r *boqRepo) AddEntry(e *domain.BOQEntry) (*domain.BOQEntry, error) {
	res, err := r.db.Exec(
		`INSERT INTO boq_entries
		 (project_id, sheet_id, item_no, dsr_item_id, description, category,
		  length, breadth, height, quantity, unit, rate, amount)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ProjectID, e.SheetID, e.ItemNo, e.DSRItemID, e.Description, e.Category,
		e.Length, e.Breadth, e.Height, e.Quantity, e.Unit, e.Rate, e.Amount,
	)
	if err != nil {
		return nil, fmt.Errorf("add boq entry: %w", err)
	}
	id, _ := res.LastInsertId()
	e.ID = id
	return e, nil
}

func (r *boqRepo) ListByProject(projectID, sheetID int64) ([]domain.BOQEntry, error) {
	rows, err := r.db.Query(
		`SELECT b.id, b.project_id, b.sheet_id, b.item_no, b.dsr_item_id, b.description, b.category,
		        b.length, b.breadth, b.height, b.quantity, b.unit, b.rate, b.amount, COALESCE(d.code, '')
		 FROM boq_entries b
		 LEFT JOIN dsr_items d ON b.dsr_item_id = d.id
		 WHERE b.project_id = ? AND b.sheet_id = ? ORDER BY b.category ASC, b.item_no ASC`, projectID, sheetID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []domain.BOQEntry
	for rows.Next() {
		var e domain.BOQEntry
		if err := rows.Scan(
			&e.ID, &e.ProjectID, &e.SheetID, &e.ItemNo, &e.DSRItemID, &e.Description, &e.Category,
			&e.Length, &e.Breadth, &e.Height, &e.Quantity, &e.Unit, &e.Rate, &e.Amount, &e.DSRItemCode,
		); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func (r *boqRepo) GetEntry(id, projectID int64) (*domain.BOQEntry, error) {
	var e domain.BOQEntry
	err := r.db.QueryRow(
		`SELECT id, project_id, sheet_id, item_no, dsr_item_id, description, category,
		        length, breadth, height, quantity, unit, rate, amount
		 FROM boq_entries WHERE id=? AND project_id=?`, id, projectID,
	).Scan(
		&e.ID, &e.ProjectID, &e.SheetID, &e.ItemNo, &e.DSRItemID, &e.Description, &e.Category,
		&e.Length, &e.Breadth, &e.Height, &e.Quantity, &e.Unit, &e.Rate, &e.Amount,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &e, err
}

func (r *boqRepo) UpdateEntry(e *domain.BOQEntry) error {
	_, err := r.db.Exec(
		`UPDATE boq_entries
		 SET length=?, breadth=?, height=?, quantity=?, rate=?, amount=?
		 WHERE id=? AND project_id=?`,
		e.Length, e.Breadth, e.Height, e.Quantity, e.Rate, e.Amount,
		e.ID, e.ProjectID,
	)
	return err
}

func (r *boqRepo) DeleteEntry(id, projectID int64) error {
	_, err := r.db.Exec(
		`DELETE FROM boq_entries WHERE id=? AND project_id=?`, id, projectID,
	)
	return err
}
