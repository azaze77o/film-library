package httpapi

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/project/omdbapp/internal/service"
)

type SearchHandler struct {
	svc    *service.SearchService
	logger *slog.Logger
}

func NewSearchHandler(svc *service.SearchService, logger *slog.Logger) *SearchHandler {
	return &SearchHandler{svc: svc, logger: logger}
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	// Достаем название фильма из параметра строки запроса
	q := r.URL.Query().Get("q")
	if q == "" {
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", "query q is required")
		return
	}

	// Определяем номер страницы из параметров строки запроса
	// По умолчанию 1 - номер страницы
	page := 1
	if raw := r.URL.Query().Get("page"); raw != "" {
		if p, err := strconv.Atoi(raw); err == nil && p > 0 {
			page = p
		}
	}

	result, err := h.svc.Search(q, page)
	if err != nil {
		handleServiceError(w, h.logger, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}
