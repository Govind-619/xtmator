package repository

import (
	"database/sql"
	"embed"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/schema.sql
var schemaFS embed.FS

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
	}
	for _, m := range migrations {
		db.Exec(m) // intentionally ignore error (column may already exist)
	}

	return db, nil
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
