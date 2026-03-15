package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Govind-619/xtmator/domain"
	"github.com/Govind-619/xtmator/usecase"
)

// ProjectHandler handles /api/projects/* routes.
type ProjectHandler struct {
	projects *usecase.ProjectUsecase
	auth     *usecase.AuthUsecase
}

func NewProjectHandler(projects *usecase.ProjectUsecase, auth *usecase.AuthUsecase) *ProjectHandler {
	return &ProjectHandler{projects: projects, auth: auth}
}

// HandleProjects routes GET/POST /api/projects
func (h *ProjectHandler) HandleProjects(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	switch r.Method {
	case http.MethodGet:
		list, err := h.projects.List(userID)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if list == nil {
			list = []domain.Project{}
		}
		jsonOK(w, list)
	case http.MethodPost:
		var req struct {
			Name       string `json:"name"`
			ClientName string `json:"client_name"`
			Location   string `json:"location"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid body", http.StatusBadRequest)
			return
		}
		p, err := h.projects.Create(userID, req.Name, req.ClientName, req.Location)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonOK(w, p)
	default:
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleProject routes GET/DELETE /api/projects/:id
func (h *ProjectHandler) HandleProject(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	// Extract :id from path /api/projects/123
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	id, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	if err != nil {
		jsonError(w, "invalid project id", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		p, err := h.projects.Get(id, userID)
		if err != nil {
			jsonError(w, err.Error(), http.StatusNotFound)
			return
		}
		jsonOK(w, p)
	case http.MethodDelete:
		if err := h.projects.Delete(id, userID); err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, map[string]string{"status": "deleted"})
	default:
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
