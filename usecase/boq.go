package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Govind-619/xtmator/domain"
	"github.com/Govind-619/xtmator/repository"
)

// BOQSheet is the full computed BOQ for a project, returned to handlers.
type BOQSheet struct {
	Project    *domain.Project
	Entries    []domain.BOQEntry
	GrandTotal float64
}

// BOQUsecase handles bill of quantities logic.
type BOQUsecase struct {
	boq      repository.BOQRepository
	dsr      repository.DSRRepository
	projects repository.ProjectRepository
	sheets   repository.ProjectSheetRepository
}

func NewBOQUsecase(boq repository.BOQRepository, dsr repository.DSRRepository, projects repository.ProjectRepository, sheets repository.ProjectSheetRepository) *BOQUsecase {
	return &BOQUsecase{boq: boq, dsr: dsr, projects: projects, sheets: sheets}
}

// dimMode returns how many dimensions are meaningful for calculating quantity.
//
//	"3d" → CUM / M3  (L × B × H)
//	"2d" → SQM / M2  (L × B)
//	"1d" → M / RMT   (L only)
//	"0d" → everything else (KG, MT, NO., LS, DAY, HR …) — manual qty required
func dimMode(unit string) string {
	switch strings.ToUpper(strings.TrimSpace(unit)) {
	case "CUM", "M3":
		return "3d"
	case "SQM", "SQM.", "M2":
		return "2d"
	case "M", "RMT":
		return "1d"
	default:
		return "0d"
	}
}

// AddItem adds one BOQ line item to a project.
// Quantity is calculated based on the item's unit:
//   - CUM/M3  → L × B × H
//   - SQM/M2  → L × B
//   - M/RMT   → L
//   - others  → manualQty
//
// Rate comes from the DSR catalogue unless manualRate > 0.
func (u *BOQUsecase) AddItem(
	projectID, sheetID int64,
	dsrItemID *int64,
	description, category, unit string,
	l, b, h float64,
	manualQty, manualRate float64,
) (*domain.BOQEntry, error) {

	// Resolve DSR item (rate, unit, description)
	rate := manualRate

	if sheetID == 0 {
		sheets, err := u.sheets.ListByProject(projectID)
		if err == nil && len(sheets) > 0 {
			sheetID = sheets[0].ID
		}
	}

	if dsrItemID != nil {
		item, err := u.dsr.GetByID(*dsrItemID)
		if err != nil {
			return nil, fmt.Errorf("fetch dsr item: %w", err)
		}
		if item == nil {
			return nil, errors.New("DSR item not found")
		}
		if description == "" {
			description = item.Description
		}
		if category == "" {
			category = item.Category
		}
		unit = item.Unit
		if rate == 0 {
			rate = item.Rate // use DSR rate only if user didn't override
		}
	}
	if unit == "" {
		unit = "CUM" // fallback for fully manual entries
	}
	if rate <= 0 {
		return nil, errors.New("rate must be greater than zero")
	}

	// Resolve quantity based on unit type
	var qty float64
	switch dimMode(unit) {
	case "3d":
		if l > 0 && b > 0 && h > 0 {
			qty = l * b * h
		}
	case "2d":
		if l > 0 && b > 0 {
			qty = l * b
		}
	case "1d":
		if l > 0 {
			qty = l
		}
	}
	// Fall back to manual qty if dimension-based calc yielded nothing
	if qty <= 0 {
		qty = manualQty
	}
	if qty <= 0 {
		return nil, errors.New("quantity must be greater than zero — provide dimensions or a manual quantity")
	}

	entry := &domain.BOQEntry{
		ProjectID:   projectID,
		SheetID:     sheetID,
		ItemNo:      u.boq.NextItemNo(projectID, sheetID),
		DSRItemID:   dsrItemID,
		Description: description,
		Category:    category,
		Length:      l,
		Breadth:     b,
		Height:      h,
		Quantity:    qty,
		Unit:        unit,
		Rate:        rate,
		Amount:      qty * rate,
	}
	return u.boq.AddEntry(entry)
}

// GetSheet returns the full BOQ sheet and grand total for a project.
func (u *BOQUsecase) GetSheet(projectID, sheetID, userID int64) (*BOQSheet, error) {
	project, err := u.projects.GetByID(projectID, userID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("project not found")
	}
	
	if sheetID == 0 {
		sheets, err := u.sheets.ListByProject(projectID)
		if err == nil && len(sheets) > 0 {
			sheetID = sheets[0].ID
		}
	}

	entries, err := u.boq.ListByProject(projectID, sheetID)
	if err != nil {
		return nil, err
	}
	var total float64
	ratio := 1.0 + (project.CostIndex / 100.0)

	for i := range entries {
		entries[i].Rate = entries[i].Rate * ratio
		entries[i].Amount = entries[i].Quantity * entries[i].Rate
		total += entries[i].Amount
	}

	return &BOQSheet{Project: project, Entries: entries, GrandTotal: total}, nil
}

// GetSharedSheet returns the full BOQ sheet and grand total using a public token.
func (u *BOQUsecase) GetSharedSheet(token string, sheetID int64) (*BOQSheet, error) {
	project, err := u.projects.GetByShareToken(token)
	if err != nil { return nil, err }
	if project == nil { return nil, errors.New("invalid share token") }

	if sheetID == 0 {
		sheets, err := u.sheets.ListByProject(project.ID)
		if err == nil && len(sheets) > 0 { sheetID = sheets[0].ID }
	}

	entries, err := u.boq.ListByProject(project.ID, sheetID)
	if err != nil { return nil, err }

	var total float64
	ratio := 1.0 + (project.CostIndex / 100.0)

	for i := range entries {
		entries[i].Rate = entries[i].Rate * ratio
		entries[i].Amount = entries[i].Quantity * entries[i].Rate
		total += entries[i].Amount
	}
	return &BOQSheet{Project: project, Entries: entries, GrandTotal: total}, nil
}

// UpdateItem updates the dimensions and rate of a single preexisting entry
func (u *BOQUsecase) UpdateItem(projectID, entryID int64, l, b, h, manualQty, manualRate float64) (*domain.BOQEntry, error) {
	entry, err := u.boq.GetEntry(entryID, projectID)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, errors.New("entry not found")
	}

	var qty float64
	switch dimMode(entry.Unit) {
	case "3d":
		if l > 0 && b > 0 && h > 0 { qty = l * b * h }
	case "2d":
		if l > 0 && b > 0 { qty = l * b }
	case "1d":
		if l > 0 { qty = l }
	}
	if qty <= 0 { qty = manualQty }
	if qty <= 0 { return nil, errors.New("quantity must be greater than zero") }

	rate := manualRate
	if rate <= 0 { rate = entry.Rate }

	entry.Length = l
	entry.Breadth = b
	entry.Height = h
	entry.Quantity = qty
	entry.Rate = rate
	entry.Amount = qty * rate

	err = u.boq.UpdateEntry(entry)
	return entry, err
}

// DeleteItem removes a BOQ entry from a project.
func (u *BOQUsecase) DeleteItem(entryID, projectID int64) error {
	return u.boq.DeleteEntry(entryID, projectID)
}
