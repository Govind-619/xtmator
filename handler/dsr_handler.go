package handler

import (
	"net/http"

	"github.com/Govind-619/xtmator/domain"
	"github.com/Govind-619/xtmator/repository"
)

// DSRHandler handles /api/dsr/* routes.
type DSRHandler struct {
	dsr repository.DSRRepository
}

func NewDSRHandler(dsr repository.DSRRepository) *DSRHandler {
	return &DSRHandler{dsr: dsr}
}

// Categories handles GET /api/dsr/categories
func (h *DSRHandler) Categories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cats, err := h.dsr.ListCategories()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if cats == nil {
		cats = []string{}
	}
	jsonOK(w, cats)
}

// Items handles GET /api/dsr/items?category=PCC
func (h *DSRHandler) Items(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	category := r.URL.Query().Get("category")
	if category == "" {
		jsonError(w, "category query param is required", http.StatusBadRequest)
		return
	}
	items, err := h.dsr.ListByCategory(category)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if items == nil {
		items = []domain.DSRItem{}
	}
	jsonOK(w, items)
}
