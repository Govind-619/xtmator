package usecase

import (
	"errors"

	"github.com/Govind-619/xtmator/domain"
	"github.com/Govind-619/xtmator/repository"
)

type CustomItemUsecase struct {
	items repository.CustomItemRepository
}

func NewCustomItemUsecase(items repository.CustomItemRepository) *CustomItemUsecase {
	return &CustomItemUsecase{items: items}
}

func (u *CustomItemUsecase) Create(userID int64, category, description, unit string, rate float64) (*domain.CustomItem, error) {
	if description == "" {
		return nil, errors.New("description is required")
	}
	if rate <= 0 {
		return nil, errors.New("rate must be greater than zero")
	}
	if unit == "" {
		unit = "CUM"
	}
	if category == "" {
		category = "Custom"
	}

	item := &domain.CustomItem{
		UserID:      userID,
		Category:    category,
		Description: description,
		Unit:        unit,
		Rate:        rate,
	}
	return u.items.Create(item)
}

func (u *CustomItemUsecase) List(userID int64) ([]domain.CustomItem, error) {
	return u.items.ListByUser(userID)
}

func (u *CustomItemUsecase) Delete(id, userID int64) error {
	return u.items.Delete(id, userID)
}
