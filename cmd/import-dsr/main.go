package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tealeg/xlsx"
)

// DSR item from parsed Excel
type dsrItem struct {
	Category    string
	Code        string
	Description string
	Unit        string
	Rate        float64
}

func main() {
	excelPath := flag.String("excel", "DSR_Analizer_Ver.18.xls", "Path to DSR analyser Excel file")
	dbPath    := flag.String("db", "xtmator.db", "Path to SQLite database file")
	category  := flag.String("category", "PCC", "Category name to assign to imported items")
	chapter   := flag.String("chapter", "4.", "Item code prefix to filter (e.g. '4.' for PCC Chapter 4)")
	flag.Parse()

	fmt.Printf("📂 Opening Excel: %s\n", *excelPath)
	wb, err := xlsx.OpenFile(*excelPath)
	if err != nil {
		log.Fatalf("Cannot open Excel: %v", err)
	}

	// Find the "DAR Original" sheet
	var sh *xlsx.Sheet
	for _, s := range wb.Sheets {
		if strings.Contains(s.Name, "DAR Original") || strings.Contains(s.Name, "DAR-Original") {
			sh = s
			break
		}
	}
	if sh == nil {
		// fallback: use first sheet
		sh = wb.Sheets[0]
		fmt.Printf("⚠  Sheet 'DAR Original' not found — using: %s\n", sh.Name)
	} else {
		fmt.Printf("✅ Using sheet: %s (%d rows)\n", sh.Name, len(sh.Rows))
	}

	// Parse rows — DAR Original structure:
	// Col 0: Sr. No (or blank), Col 1: Item Code, Col 3: Description, Col 4: Unit, Col 5: Rate (basic), Col 6: Rate (total w/ overhead)
	var items []dsrItem
	for _, row := range sh.Rows {
		if len(row.Cells) < 5 {
			continue
		}
		code := strings.TrimSpace(row.Cells[1].String())
		if code == "" || !strings.HasPrefix(code, *chapter) {
			continue
		}
		// Find description — it may span multiple cells
		desc := ""
		for ci := 2; ci < len(row.Cells) && ci <= 4; ci++ {
			v := strings.TrimSpace(row.Cells[ci].String())
			if v != "" && !isNumeric(v) {
				desc = v
				break
			}
		}
		if desc == "" {
			continue
		}

		unit := ""
		var rate float64
		// Scan remaining cells for unit and rate
		for ci := 3; ci < len(row.Cells); ci++ {
			v := strings.TrimSpace(row.Cells[ci].String())
			if isUnit(v) && unit == "" {
				unit = v
			} else if isNumeric(v) && rate == 0 {
				fmt.Sscanf(v, "%f", &rate)
			}
		}
		// Prefer the "total rate" column (col 6 if available)
		if len(row.Cells) > 6 {
			v := strings.TrimSpace(row.Cells[6].String())
			var r float64
			if n, _ := fmt.Sscanf(v, "%f", &r); n == 1 && r > 0 {
				rate = r
			}
		}
		if rate == 0 {
			continue
		}
		if unit == "" {
			unit = "CUM"
		}
		// Clean description (remove double-spaces, newlines)
		desc = strings.Join(strings.Fields(desc), " ")

		items = append(items, dsrItem{
			Category:    *category,
			Code:        code,
			Description: desc,
			Unit:        unit,
			Rate:        rate,
		})
	}

	fmt.Printf("🔍 Found %d items with prefix '%s'\n", len(items), *chapter)
	if len(items) == 0 {
		fmt.Println("⚠  No items found. Check --chapter flag matches your DSR item code prefix.")
		return
	}

	// Open DB and insert
	db, err := sql.Open("sqlite3", *dbPath+"?_foreign_keys=on")
	if err != nil {
		log.Fatalf("Open DB: %v", err)
	}
	defer db.Close()

	tx, _ := db.Begin()
	stmt, err := tx.Prepare(`
		INSERT INTO dsr_items (category, code, description, unit, rate)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(code) DO UPDATE SET description=excluded.description, rate=excluded.rate, unit=excluded.unit
	`)
	if err != nil {
		log.Fatalf("Prepare: %v", err)
	}

	imported := 0
	for _, item := range items {
		if _, err := stmt.Exec(item.Category, item.Code, item.Description, item.Unit, item.Rate); err != nil {
			fmt.Printf("  ⚠  Skip [%s]: %v\n", item.Code, err)
			continue
		}
		fmt.Printf("  ✅ [%s] %s — ₹%.2f/%s\n", item.Code, truncate(item.Description, 50), item.Rate, item.Unit)
		imported++
	}
	tx.Commit()

	fmt.Printf("\n🎉 Imported %d DSR items into %s\n", imported, *dbPath)
}

func isNumeric(s string) bool {
	s = strings.ReplaceAll(strings.ReplaceAll(s, ",", ""), " ", "")
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return err == nil
}

func isUnit(s string) bool {
	units := []string{"Cum", "CUM", "Sqm", "SQM", "Sqm.", "Rmt", "Kg", "MT", "No.", "No", "L.S.", "Day", "Hr", "Month", "m3", "m2"}
	su := strings.TrimSpace(s)
	for _, u := range units {
		if strings.EqualFold(su, u) {
			return true
		}
	}
	return false
}

func truncate(s string, n int) string {
	if len(s) <= n { return s }
	return s[:n] + "…"
}
