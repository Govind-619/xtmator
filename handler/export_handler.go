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
			col.New(1).WithStyle(cStyle).Add(text.New(fmt.Sprintf("%.2f", e.Rate), numStyle)),
			col.New(1).WithStyle(cStyle).Add(text.New(fmt.Sprintf("%.2f", e.Amount), numStyle)),
		)
		sr++
	}

	m.AddRow(4) // spacer

	sumLabelStyle := props.Text{Size: 10, Style: fontstyle.Bold, Align: align.Right, Top: 3, Bottom: 3, Left: 2, Right: 2}
	sumValStyle := props.Text{Size: 10, Style: fontstyle.Bold, Align: align.Right, Top: 3, Bottom: 3, Left: 2, Right: 2}

	// Grand total
	m.AddRows(row.New(12).Add(
		col.New(10).WithStyle(cStyle).Add(text.New("GRAND TOTAL", sumLabelStyle)),
		col.New(2).WithStyle(cStyle).Add(text.New(fmt.Sprintf("Rs. %.2f", sheet.GrandTotal), sumValStyle)),
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
			col.New(2).WithStyle(cStyle).Add(text.New(fmt.Sprintf("%.2f", catSums[cat]), props.Text{
				Size: 9, Style: fontstyle.Bold, Align: align.Right, Top: 2, Bottom: 2, Left: 2, Right: 2,
			})),
		))
	}

	summaryPage.Add(row.New(6)) // spacer

	// Add Grand Total to the bottom of the summary table
	summaryPage.Add(row.New(12).Add(
		col.New(10).WithStyle(cStyle).Add(text.New("GRAND TOTAL", sumLabelStyle)),
		col.New(2).WithStyle(cStyle).Add(text.New(fmt.Sprintf("Rs. %.2f", sheet.GrandTotal), sumValStyle)),
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
