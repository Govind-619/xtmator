package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Govind-619/xtmator/usecase"
)

// BOQHandler handles /api/projects/:id/boq/* routes.
type BOQHandler struct {
	boq  *usecase.BOQUsecase
	auth *usecase.AuthUsecase
}

func NewBOQHandler(boq *usecase.BOQUsecase, auth *usecase.AuthUsecase) *BOQHandler {
	return &BOQHandler{boq: boq, auth: auth}
}

// HandleBOQ routes GET/POST /api/projects/:id/boq
func (h *BOQHandler) HandleBOQ(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	projectID := extractProjectID(r.URL.Path)
	if projectID == 0 {
		jsonError(w, "invalid project id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		sheetID, _ := strconv.ParseInt(r.URL.Query().Get("sheet_id"), 10, 64)
		sheet, err := h.boq.GetSheet(projectID, sheetID, userID)
		if err != nil {
			jsonError(w, err.Error(), http.StatusNotFound)
			return
		}
		jsonOK(w, map[string]interface{}{
			"project":     sheet.Project,
			"entries":     sheet.Entries,
			"grand_total": sheet.GrandTotal,
		})

	case http.MethodPost:
		var req struct {
			SheetID     int64   `json:"sheet_id"`
			DSRItemID   *int64  `json:"dsr_item_id"`
			Description string  `json:"description"`
			Category    string  `json:"category"`
			Unit        string  `json:"unit"`
			Length      float64 `json:"length"`
			Breadth     float64 `json:"breadth"`
			Height      float64 `json:"height"`
			ManualQty   float64 `json:"manual_qty"`
			ManualRate  float64 `json:"manual_rate"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid body", http.StatusBadRequest)
			return
		}
		entry, err := h.boq.AddItem(
			projectID, req.SheetID,
			req.DSRItemID,
			req.Description, req.Category, req.Unit,
			req.Length, req.Breadth, req.Height,
			req.ManualQty, req.ManualRate,
		)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonOK(w, entry)

	default:
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleBOQEntry handles DELETE and PUT /api/projects/:id/boq/:entryId
func (h *BOQHandler) HandleBOQEntry(w http.ResponseWriter, r *http.Request) {
	projectID := extractProjectID(r.URL.Path)
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	entryID, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	if err != nil || projectID == 0 {
		jsonError(w, "invalid ids", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		if err := h.boq.DeleteItem(entryID, projectID); err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, map[string]string{"status": "deleted"})
	case http.MethodPut:
		var req struct {
			Length     float64 `json:"length"`
			Breadth    float64 `json:"breadth"`
			Height     float64 `json:"height"`
			ManualQty  float64 `json:"manual_qty"`
			ManualRate float64 `json:"manual_rate"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid body", http.StatusBadRequest)
			return
		}
		entry, err := h.boq.UpdateItem(projectID, entryID, req.Length, req.Breadth, req.Height, req.ManualQty, req.ManualRate)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonOK(w, entry)
	default:
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// extractProjectID pulls the numeric project ID from paths like /api/projects/42/boq
func extractProjectID(path string) int64 {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	// parts: ["api","projects","42","boq",...]
	for i, p := range parts {
		if p == "projects" && i+1 < len(parts) {
			id, _ := strconv.ParseInt(parts[i+1], 10, 64)
			return id
		}
	}
	return 0
}
