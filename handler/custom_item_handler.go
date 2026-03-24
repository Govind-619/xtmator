package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Govind-619/xtmator/domain"
	"github.com/Govind-619/xtmator/usecase"
)

type CustomItemHandler struct {
	items *usecase.CustomItemUsecase
	auth  *usecase.AuthUsecase
}

func NewCustomItemHandler(items *usecase.CustomItemUsecase, auth *usecase.AuthUsecase) *CustomItemHandler {
	return &CustomItemHandler{items: items, auth: auth}
}

func (h *CustomItemHandler) HandleItems(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	switch r.Method {
	case http.MethodGet:
		list, err := h.items.List(userID)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if list == nil {
			list = []domain.CustomItem{}
		}
		jsonOK(w, list)

	case http.MethodPost:
		var req struct {
			Category    string  `json:"category"`
			Description string  `json:"description"`
			Unit        string  `json:"unit"`
			Rate        float64 `json:"rate"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid body", http.StatusBadRequest)
			return
		}
		item, err := h.items.Create(userID, req.Category, req.Description, req.Unit, req.Rate)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonOK(w, item)

	default:
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *CustomItemHandler) HandleItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserID(r)
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	id, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	if err != nil {
		jsonError(w, "invalid id format", http.StatusBadRequest)
		return
	}
	if err := h.items.Delete(id, userID); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]string{"status": "deleted"})
}
