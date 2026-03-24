package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Govind-619/xtmator/usecase"
)

type ShareHandler struct {
	boq *usecase.BOQUsecase
}

func NewShareHandler(boq *usecase.BOQUsecase) *ShareHandler {
	return &ShareHandler{boq: boq}
}

func (h *ShareHandler) HandleSharedSheet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var token string
	for i, p := range parts {
		if p == "share" && i+1 < len(parts) {
			token = parts[i+1]
			break
		}
	}
	if token == "" {
		jsonError(w, "invalid token", http.StatusBadRequest)
		return
	}

	sheetID, _ := strconv.ParseInt(r.URL.Query().Get("sheet_id"), 10, 64)

	sheet, err := h.boq.GetSharedSheet(token, sheetID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}

	jsonOK(w, map[string]interface{}{
		"project":     sheet.Project,
		"entries":     sheet.Entries,
		"grand_total": sheet.GrandTotal,
	})
}
