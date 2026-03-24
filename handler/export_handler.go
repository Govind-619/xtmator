package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Govind-619/xtmator/usecase"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/page"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/xuri/excelize/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

// ExportHandler handles PDF export for a project BOQ.
type ExportHandler struct {
	boq  *usecase.BOQUsecase
	auth *usecase.AuthUsecase
}

func NewExportHandler(boq *usecase.BOQUsecase, auth *usecase.AuthUsecase) *ExportHandler {
	return &ExportHandler{boq: boq, auth: auth}
}

// ExportPDF handles GET /api/projects/:id/export/pdf
func (h *ExportHandler) ExportPDF(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserID(r)
	projectID := extractProjectID(r.URL.Path)
	if projectID == 0 {
		jsonError(w, "invalid project id", http.StatusBadRequest)
		return
	}

	user, err := h.auth.GetUser(userID)
	if err != nil || user == nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	sheet, err := h.boq.GetSheet(projectID, userID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}

	cfg := config.NewBuilder().
		WithPageNumber(props.PageNumber{
			Pattern: "Page {current} of {total}",
			Place:   props.RightBottom,
		}).
		WithOrientation(orientation.Horizontal).
		WithLeftMargin(15).
		WithRightMargin(15).
		WithTopMargin(15).
		Build()

	m := maroto.New(cfg)

	// App Name
	m.AddRows(row.New(8).Add(
		col.New(12).Add(text.New("XTMATOR", props.Text{
			Size:  12,
			Style: fontstyle.Bold,
			Align: align.Center,
		})),
	))
	m.AddRow(2) // spacer

	// Title
	m.AddRows(row.New(14).Add(
		col.New(12).Add(text.New("BILL OF QUANTITIES", props.Text{
			Size:  16,
			Style: fontstyle.Bold,
			Align: align.Center,
		})),
	))

	// Project info
	m.AddRows(row.New(8).Add(
		col.New(12).Add(text.New(
			fmt.Sprintf("Project: %s  |  Client: %s  |  Location: %s",
				sheet.Project.Name, sheet.Project.ClientName, sheet.Project.Location),
			props.Text{Size: 9, Align: align.Center},
		)),
	))

	m.AddRow(4) // spacer

	// Table header
	m.AddRows(headerRow())

	cStyle := &props.Cell{BorderType: border.Full}
	txtStyle := props.Text{Size: 8, Align: align.Left, Top: 2, Bottom: 2, Left: 2, Right: 2}
	numStyle := props.Text{Size: 8, Align: align.Right, Top: 2, Bottom: 2, Left: 2, Right: 2}

	// Group by category to prepare for summary
	catSums := make(map[string]float64)
	var cats []string
	
	// Track state for grouping
	var lastCat string
	sr := 1

	// BOQ rows
	for _, e := range sheet.Entries {
		if _, ok := catSums[e.Category]; !ok {
			cats = append(cats, e.Category)
		}
		catSums[e.Category] += e.Amount

		if e.Category != lastCat {
			// Print a category header row that spans exactly the 12 columns
			// 1+4+1+1+1+1+1+1+1 = 12 columns
			catHeaderStyle := props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Left, Top: 3, Bottom: 3, Left: 4, Right: 4}
			m.AddRows(row.New(10).Add(
				col.New(12).WithStyle(cStyle).Add(text.New(e.Category, catHeaderStyle)),
			))
			lastCat = e.Category
		}

		desc := e.Description
		if e.DSRItemCode != "" && !strings.Contains(e.Description, "["+e.DSRItemCode+"]") {
			desc = fmt.Sprintf("[%s] %s", e.DSRItemCode, e.Description)
		}
		m.AddAutoRow(
			col.New(1).WithStyle(cStyle).Add(text.New(fmt.Sprintf("%d", sr), txtStyle)),
			col.New(4).WithStyle(cStyle).Add(text.New(desc, txtStyle)),
			col.New(1).WithStyle(cStyle).Add(text.New(e.Category, txtStyle)),
			col.New(1).WithStyle(cStyle).Add(text.New(fmt.Sprintf("%.2f", e.Length), numStyle)),
			col.New(1).WithStyle(cStyle).Add(text.New(fmt.Sprintf("%.2f", e.Breadth), numStyle)),
			col.New(1).WithStyle(cStyle).Add(text.New(fmt.Sprintf("%.2f", e.Height), numStyle)),
			col.New(1).WithStyle(cStyle).Add(text.New(fmt.Sprintf("%.3f", e.Quantity), numStyle)),
			col.New(1).WithStyle(cStyle).Add(text.New(formatIndianCurrency(e.Rate), numStyle)),
			col.New(1).WithStyle(cStyle).Add(text.New(formatIndianCurrency(e.Amount), numStyle)),
		)
		sr++
	}

	m.AddRow(4) // spacer

	sumLabelStyle := props.Text{Size: 10, Style: fontstyle.Bold, Align: align.Right, Top: 3, Bottom: 3, Left: 2, Right: 2}
	sumValStyle := props.Text{Size: 10, Style: fontstyle.Bold, Align: align.Right, Top: 3, Bottom: 3, Left: 2, Right: 2}

	// Grand total
	m.AddRows(row.New(12).Add(
		col.New(10).WithStyle(cStyle).Add(text.New("GRAND TOTAL", sumLabelStyle)),
		col.New(2).WithStyle(cStyle).Add(text.New("Rs. "+formatIndianCurrency(sheet.GrandTotal), sumValStyle)),
	))

	m.AddRow(8) // spacer

	// Start new page for Summarized Categories
	summaryPage := page.New()

	// Category Summary Table Header
	summaryPage.Add(row.New(12).Add(
		col.New(12).Add(text.New("SUMMARIZED CATEGORY", props.Text{
			Size:  12,
			Style: fontstyle.Bold,
			Align: align.Center,
		})),
	))
	
	summaryPage.Add(row.New(4)) // spacer

	// Table Header for Summary Table
	summaryPage.Add(row.New(10).Add(
		col.New(1).WithStyle(cStyle).Add(text.New("Sr. No.", txtStyle)),
		col.New(9).WithStyle(cStyle).Add(text.New("Category", txtStyle)),
		col.New(2).WithStyle(cStyle).Add(text.New("Amount (Rs.)", numStyle)),
	))

	// Category Summary Table
	for i, cat := range cats {
		summaryPage.Add(row.New(10).Add(
			col.New(1).WithStyle(cStyle).Add(text.New(fmt.Sprintf("%d", i+1), txtStyle)),
			col.New(9).WithStyle(cStyle).Add(text.New(cat, props.Text{
				Size: 9, Style: fontstyle.Bold, Align: align.Left, Top: 2, Bottom: 2, Left: 2, Right: 2,
			})),
			col.New(2).WithStyle(cStyle).Add(text.New(formatIndianCurrency(catSums[cat]), props.Text{
				Size: 9, Style: fontstyle.Bold, Align: align.Right, Top: 2, Bottom: 2, Left: 2, Right: 2,
			})),
		))
	}

	summaryPage.Add(row.New(6)) // spacer

	// Add Grand Total to the bottom of the summary table
	summaryPage.Add(row.New(12).Add(
		col.New(10).WithStyle(cStyle).Add(text.New("GRAND TOTAL", sumLabelStyle)),
		col.New(2).WithStyle(cStyle).Add(text.New("Rs. "+formatIndianCurrency(sheet.GrandTotal), sumValStyle)),
	))

	summaryPage.Add(row.New(12)) // spacer

	// Footer: User stamp
	summaryPage.Add(row.New(10).Add(
		col.New(12).Add(text.New("Created by: "+user.Name, props.Text{
			Size:  9,
			Style: fontstyle.Italic,
			Align: align.Right,
		})),
	))

	m.AddPages(summaryPage)

	doc, err := m.Generate()
	if err != nil {
		jsonError(w, "pdf generation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("BOQ_%s.pdf", sanitizeFilename(sheet.Project.Name))
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Write(doc.GetBytes())
}

func headerRow() core.Row {
	txtStyle := props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Left, Top: 2, Bottom: 2, Left: 2, Right: 2}
	numStyle := props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Right, Top: 2, Bottom: 2, Left: 2, Right: 2}
	cStyle := &props.Cell{BorderType: border.Full}
	return row.New(12).Add(
		col.New(1).WithStyle(cStyle).Add(text.New("Sr.", txtStyle)),
		col.New(4).WithStyle(cStyle).Add(text.New("Description of Work", txtStyle)),
		col.New(1).WithStyle(cStyle).Add(text.New("Category", txtStyle)),
		col.New(1).WithStyle(cStyle).Add(text.New("L (m)", numStyle)),
		col.New(1).WithStyle(cStyle).Add(text.New("B (m)", numStyle)),
		col.New(1).WithStyle(cStyle).Add(text.New("H (m)", numStyle)),
		col.New(1).WithStyle(cStyle).Add(text.New("Qty (CUM)", numStyle)),
		col.New(1).WithStyle(cStyle).Add(text.New("Rate (Rs.)", numStyle)),
		col.New(1).WithStyle(cStyle).Add(text.New("Amount (Rs.)", numStyle)),
	)
}

func sanitizeFilename(s string) string {
	out := make([]byte, 0, len(s))
	for _, c := range []byte(s) {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
			out = append(out, c)
		} else {
			out = append(out, '_')
		}
	}
	return string(out)
}

func formatIndianCurrency(val float64) string {
	s := fmt.Sprintf("%.2f", val)
	parts := strings.Split(s, ".")
	intPart := parts[0]
	n := len(intPart)
	if n <= 3 {
		return s
	}
	res := intPart[n-3:]
	intPart = intPart[:n-3]
	for len(intPart) > 0 {
		if len(intPart) > 2 {
			res = intPart[len(intPart)-2:] + "," + res
			intPart = intPart[:len(intPart)-2]
		} else {
			res = intPart + "," + res
			break
		}
	}
	return res + "." + parts[1]
}

func (h *ExportHandler) ExportExcel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserID(r)
	projectID := extractProjectID(r.URL.Path)
	if projectID == 0 {
		jsonError(w, "invalid project id", http.StatusBadRequest)
		return
	}

	user, err := h.auth.GetUser(userID)
	if err != nil || user == nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	sheet, err := h.boq.GetSheet(projectID, userID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	f.SetSheetName("Sheet1", "BOQ")

	// Main BOQ Sheet Headers
	headers := []interface{}{"Sr.", "Description of Work", "Category", "L (m)", "B (m)", "H (m)", "Qty", "Rate (Rs.)", "Amount (Rs.)"}
	f.SetSheetRow("BOQ", "A1", &headers)

	// Make headers bold
	styleBold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	f.SetRowStyle("BOQ", 1, 1, styleBold)

	catSums := make(map[string]float64)
	var cats []string
	var lastCat string
	sr := 1
	rowIdx := 2

	for _, e := range sheet.Entries {
		if _, ok := catSums[e.Category]; !ok {
			cats = append(cats, e.Category)
		}
		catSums[e.Category] += e.Amount

		if e.Category != lastCat {
			// Category row spanning 9 columns visually, just bolding column B 
			f.SetCellValue("BOQ", fmt.Sprintf("B%d", rowIdx), "Category: "+e.Category)
			f.SetCellStyle("BOQ", fmt.Sprintf("B%d", rowIdx), fmt.Sprintf("B%d", rowIdx), styleBold)
			lastCat = e.Category
			rowIdx++
		}

		desc := e.Description
		if e.DSRItemCode != "" && !strings.Contains(e.Description, "["+e.DSRItemCode+"]") {
			desc = fmt.Sprintf("[%s] %s", e.DSRItemCode, e.Description)
		}

		l := ""
		if e.Length > 0 {
			l = fmt.Sprintf("%.2f", e.Length)
		}
		b := ""
		if e.Breadth > 0 {
			b = fmt.Sprintf("%.2f", e.Breadth)
		}
		height := ""
		if e.Height > 0 {
			height = fmt.Sprintf("%.2f", e.Height)
		}

		row := []interface{}{
			sr,
			desc,
			e.Category,
			l,
			b,
			height,
			fmt.Sprintf("%.3f %s", e.Quantity, e.Unit),
			e.Rate,
			e.Amount,
		}
		f.SetSheetRow("BOQ", fmt.Sprintf("A%d", rowIdx), &row)
		sr++
		rowIdx++
	}

	rowIdx++
	f.SetCellValue("BOQ", fmt.Sprintf("H%d", rowIdx), "GRAND TOTAL")
	f.SetCellValue("BOQ", fmt.Sprintf("I%d", rowIdx), sheet.GrandTotal)
	f.SetCellStyle("BOQ", fmt.Sprintf("H%d", rowIdx), fmt.Sprintf("I%d", rowIdx), styleBold)

	rowIdx += 2
	f.SetCellValue("BOQ", fmt.Sprintf("B%d", rowIdx), "Created by: "+user.Name)

	// Summary Sheet
	f.NewSheet("Summarized")
	sumHeaders := []interface{}{"Sr. No.", "Category", "Amount (Rs.)"}
	f.SetSheetRow("Summarized", "A1", &sumHeaders)
	f.SetRowStyle("Summarized", 1, 1, styleBold)

	sumRowIdx := 2
	for i, cat := range cats {
		f.SetSheetRow("Summarized", fmt.Sprintf("A%d", sumRowIdx), &[]interface{}{i + 1, cat, catSums[cat]})
		sumRowIdx++
	}

	sumRowIdx++
	f.SetCellValue("Summarized", fmt.Sprintf("B%d", sumRowIdx), "GRAND TOTAL")
	f.SetCellValue("Summarized", fmt.Sprintf("C%d", sumRowIdx), sheet.GrandTotal)
	f.SetCellStyle("Summarized", fmt.Sprintf("B%d", sumRowIdx), fmt.Sprintf("C%d", sumRowIdx), styleBold)

	f.SetActiveSheet(0)

	filename := fmt.Sprintf("BOQ_%s.xlsx", sanitizeFilename(sheet.Project.Name))
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	f.Write(w)
}
