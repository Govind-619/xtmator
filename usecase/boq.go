package usecase

import (
	"errors"
	"fmt"

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

// AddItem adds one BOQ line item to a project.
// If dsrItemID is provided, rate is fetched from the DSR catalogue (override if manualRate > 0).
// Quantity = L × B × H if all three dims are > 0, otherwise manualQty is used.
func (u *BOQUsecase) AddItem(
	projectID int64,
	dsrItemID *int64,
	description, category string,
	l, b, h float64,
	manualQty, manualRate float64,
) (*domain.BOQEntry, error) {

	// Resolve rate
	rate := manualRate
	unit := "CUM"
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
			rate = item.Rate // only use DSR rate if user didn't override
		}
	}
	if rate <= 0 {
		return nil, errors.New("rate must be greater than zero")
	}

	// Resolve quantity
	qty := manualQty
	if l > 0 && b > 0 && h > 0 {
		qty = l * b * h
	}
	if qty <= 0 {
		return nil, errors.New("quantity must be greater than zero — provide L×B×H or a manual quantity")
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
