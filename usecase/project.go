package usecase

import (
	"errors"

	"github.com/Govind-619/xtmator/domain"
	"github.com/Govind-619/xtmator/repository"
)

// ProjectUsecase handles project management for a user.
type ProjectUsecase struct {
	projects repository.ProjectRepository
}

func NewProjectUsecase(projects repository.ProjectRepository) *ProjectUsecase {
	return &ProjectUsecase{projects: projects}
}

func (u *ProjectUsecase) Create(userID int64, name, clientName, location string) (*domain.Project, error) {
	if name == "" {
		return nil, errors.New("project name is required")
	}
	return u.projects.Create(userID, name, clientName, location)
}

func (u *ProjectUsecase) List(userID int64) ([]domain.Project, error) {
	return u.projects.ListByUser(userID)
}

func (u *ProjectUsecase) Get(id, userID int64) (*domain.Project, error) {
	p, err := u.projects.GetByID(id, userID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errors.New("project not found")
	}
	return p, nil
}

func (u *ProjectUsecase) Delete(id, userID int64) error {
	return u.projects.Delete(id, userID)
}
