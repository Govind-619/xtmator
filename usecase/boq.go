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
}

func NewBOQUsecase(boq repository.BOQRepository, dsr repository.DSRRepository, projects repository.ProjectRepository) *BOQUsecase {
	return &BOQUsecase{boq: boq, dsr: dsr, projects: projects}
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
	projectID int64,
	dsrItemID *int64,
	description, category string,
	l, b, h float64,
	manualQty, manualRate float64,
) (*domain.BOQEntry, error) {

	// Resolve DSR item (rate, unit, description)
	rate := manualRate
	unit := ""

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
		ItemNo:      u.boq.NextItemNo(projectID),
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
func (u *BOQUsecase) GetSheet(projectID, userID int64) (*BOQSheet, error) {
	project, err := u.projects.GetByID(projectID, userID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("project not found")
	}
	entries, err := u.boq.ListByProject(projectID)
	if err != nil {
		return nil, err
	}
	var total float64
	for _, e := range entries {
		total += e.Amount
	}
	return &BOQSheet{Project: project, Entries: entries, GrandTotal: total}, nil
}

// DeleteItem removes a BOQ entry from a project.
func (u *BOQUsecase) DeleteItem(entryID, projectID int64) error {
	return u.boq.DeleteEntry(entryID, projectID)
}
