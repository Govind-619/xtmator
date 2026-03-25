package repository

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/schema.sql
var schemaFS embed.FS

//go:embed migrations/dsr_seed.json
var dsrSeedData []byte

// NewSQLiteDB opens (or creates) the SQLite database, applies schema, and runs
// safe ALTER TABLE migrations for columns added after initial schema.
func NewSQLiteDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	// Apply main schema
	schema, err := schemaFS.ReadFile("migrations/schema.sql")
	if err != nil {
		return nil, fmt.Errorf("read schema: %w", err)
	}
	for _, stmt := range splitSQL(string(schema)) {
		if _, err := db.Exec(stmt); err != nil {
			return nil, fmt.Errorf("exec schema [%.40s]: %w", stmt, err)
		}
	}

	// Safe incremental migrations — ignore error if column already exists
	migrations := []string{
		`ALTER TABLE users ADD COLUMN google_id TEXT`,
		`ALTER TABLE users ADD COLUMN auth_provider TEXT NOT NULL DEFAULT 'email'`,
		`ALTER TABLE projects ADD COLUMN cost_index REAL DEFAULT 0`,
		`CREATE TABLE IF NOT EXISTS project_sheets (
             id INTEGER PRIMARY KEY AUTOINCREMENT,
             project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
             name TEXT NOT NULL DEFAULT 'Main',
             created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
             updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
         )`,
		`ALTER TABLE boq_entries ADD COLUMN sheet_id INTEGER REFERENCES project_sheets(id) ON DELETE CASCADE`,
		`INSERT INTO project_sheets (project_id, name) SELECT id, 'Main' FROM projects WHERE id NOT IN (SELECT project_id FROM project_sheets)`,
		`UPDATE boq_entries SET sheet_id = (SELECT id FROM project_sheets WHERE project_sheets.project_id = boq_entries.project_id LIMIT 1) WHERE sheet_id IS NULL`,
		`ALTER TABLE projects ADD COLUMN share_token TEXT`,
	}
	for _, m := range migrations {
		db.Exec(m) // intentionally ignore error (column may already exist)
	}

	// Internal Seeding: If dsr_items is empty, populate from embedded JSON
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM dsr_items").Scan(&count)
	if err == nil && count == 0 {
		fmt.Println("🌱 Seeding DSR items from embedded catalog...")
		if err := seedDSR(db); err != nil {
			fmt.Printf("⚠️  Warning: Failed to seed DSR items: %v\n", err)
		} else {
			fmt.Println("✅ DSR seeding complete.")
		}
	}

	return db, nil
}

type dsrSeedItem struct {
	Code        string  `json:"code"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Unit        string  `json:"unit"`
	Rate        float64 `json:"rate"`
}

func seedDSR(db *sql.DB) error {
	var items []dsrSeedItem
	if err := json.Unmarshal(dsrSeedData, &items); err != nil {
		return fmt.Errorf("unmarshal dsr seed: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO dsr_items (code, category, description, unit, rate) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, it := range items {
		if _, err := stmt.Exec(it.Code, it.Category, it.Description, it.Unit, it.Rate); err != nil {
			return fmt.Errorf("insert dsr item %s: %w", it.Code, err)
		}
	}

	return tx.Commit()
}

func splitSQL(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ";") {
		part = strings.TrimSpace(part)
		var meaningful []string
		for _, line := range strings.Split(part, "\n") {
			t := strings.TrimSpace(line)
			if t != "" && !strings.HasPrefix(t, "--") {
				meaningful = append(meaningful, line)
			}
		}
		if len(meaningful) > 0 {
			out = append(out, strings.Join(meaningful, "\n"))
		}
	}
	return out
}

// func min(a, b int) int {
// 	if a < b {
// 		return a
// 	}
// 	return b
// }
