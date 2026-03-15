package handler

import (
	"fmt"
	"net/http"

	"github.com/Govind-619/xtmator/usecase"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
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
		WithLeftMargin(15).
		WithRightMargin(15).
		WithTopMargin(15).
		Build()

	m := maroto.New(cfg)

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

	// BOQ rows
	for _, e := range sheet.Entries {
		m.AddRows(row.New(8).Add(
			col.New(1).Add(text.New(fmt.Sprintf("%d", e.ItemNo), props.Text{Size: 8, Align: align.Center})),
			col.New(4).Add(text.New(e.Description, props.Text{Size: 8})),
			col.New(1).Add(text.New(e.Category, props.Text{Size: 8, Align: align.Center})),
			col.New(1).Add(text.New(fmt.Sprintf("%.2f", e.Length), props.Text{Size: 8, Align: align.Center})),
			col.New(1).Add(text.New(fmt.Sprintf("%.2f", e.Breadth), props.Text{Size: 8, Align: align.Center})),
			col.New(1).Add(text.New(fmt.Sprintf("%.2f", e.Height), props.Text{Size: 8, Align: align.Center})),
			col.New(1).Add(text.New(fmt.Sprintf("%.3f", e.Quantity), props.Text{Size: 8, Align: align.Right})),
			col.New(1).Add(text.New(fmt.Sprintf("%.2f", e.Rate), props.Text{Size: 8, Align: align.Right})),
			col.New(1).Add(text.New(fmt.Sprintf("%.2f", e.Amount), props.Text{Size: 8, Align: align.Right})),
		))
	}

	m.AddRow(4) // spacer

	// Grand total
	m.AddRows(row.New(10).Add(
		col.New(10).Add(text.New("GRAND TOTAL", props.Text{
			Size: 10, Style: fontstyle.Bold, Align: align.Right,
		})),
		col.New(2).Add(text.New(fmt.Sprintf("₹ %.2f", sheet.GrandTotal), props.Text{
			Size: 10, Style: fontstyle.Bold, Align: align.Right,
		})),
	))

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
	style := props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Center}
	return row.New(10).Add(
		col.New(1).Add(text.New("Sr.", style)),
		col.New(4).Add(text.New("Description of Work", style)),
		col.New(1).Add(text.New("Category", style)),
		col.New(1).Add(text.New("L (m)", style)),
		col.New(1).Add(text.New("B (m)", style)),
		col.New(1).Add(text.New("H (m)", style)),
		col.New(1).Add(text.New("Qty (CUM)", style)),
		col.New(1).Add(text.New("Rate (₹)", style)),
		col.New(1).Add(text.New("Amount (₹)", style)),
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
