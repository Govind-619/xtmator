package usecase

import (
	"crypto/rand"
	"encoding/hex"
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

func (u *ProjectUsecase) GetByShareToken(token string) (*domain.Project, error) {
	if token == "" {
		return nil, errors.New("invalid token")
	}
	return u.projects.GetByShareToken(token)
}

func (u *ProjectUsecase) GenerateShareToken(id, userID int64) (string, error) {
	p, err := u.projects.GetByID(id, userID)
	if err != nil { return "", err }
	if p == nil { return "", errors.New("project not found") }

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	p.ShareToken = token
	
	if err := u.projects.Update(p); err != nil {
		return "", err
	}
	return token, nil
}

func (u *ProjectUsecase) Update(p *domain.Project) error {
	return u.projects.Update(p)
}

func (u *ProjectUsecase) Delete(id, userID int64) error {
	return u.projects.Delete(id, userID)
}
