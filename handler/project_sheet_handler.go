package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Govind-619/xtmator/domain"
	"github.com/Govind-619/xtmator/repository"
)

type ProjectSheetHandler struct {
	sheets repository.ProjectSheetRepository
}

func NewProjectSheetHandler(sheets repository.ProjectSheetRepository) *ProjectSheetHandler {
	return &ProjectSheetHandler{sheets: sheets}
}

func (h *ProjectSheetHandler) HandleSheets(w http.ResponseWriter, r *http.Request) {
	projectID := extractProjectID(r.URL.Path)
	if projectID == 0 {
		jsonError(w, "invalid project id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		list, err := h.sheets.ListByProject(projectID)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if list == nil {
			list = []domain.ProjectSheet{}
		}
		jsonOK(w, list)
	case http.MethodPost:
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid body", http.StatusBadRequest)
			return
		}
		if req.Name == "" {
			req.Name = "New Sheet"
		}
		s, err := h.sheets.Create(projectID, req.Name)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, s)
	default:
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ProjectSheetHandler) HandleSheet(w http.ResponseWriter, r *http.Request) {
	projectID := extractProjectID(r.URL.Path)
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	sheetID, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	if err != nil || projectID == 0 {
		jsonError(w, "invalid ids", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPut:
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid body", http.StatusBadRequest)
			return
		}
		s, err := h.sheets.GetByID(sheetID, projectID)
		if err != nil || s == nil {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		s.Name = req.Name
		if err := h.sheets.Update(s); err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, s)
	case http.MethodDelete:
		if err := h.sheets.Delete(sheetID, projectID); err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, map[string]string{"status": "deleted"})
	default:
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
