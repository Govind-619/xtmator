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

// NewSQLiteDB opens (or creates) the SQLite database, applies schema, and returns the db handle.
func NewSQLiteDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	schema, err := schemaFS.ReadFile("migrations/schema.sql")
	if err != nil {
		return nil, fmt.Errorf("read schema: %w", err)
	}
	for _, stmt := range splitSQL(string(schema)) {
		if _, err := db.Exec(stmt); err != nil {
			return nil, fmt.Errorf("exec schema [%s]: %w", stmt[:min(40, len(stmt))], err)
		}
	}
	return db, nil
}

func splitSQL(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ";") {
		part = strings.TrimSpace(part)
		// skip blank lines and comment-only blocks
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
