package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/project/omdbapp/internal/domain"
)

// handleServiceError — единая точка сопоставления доменных ошибок с HTTP-кодами.
// Handler'ы просто вызывают её и не думают про коды самостоятельно.

func handleServiceError(w http.ResponseWriter, logger *slog.Logger, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "resource not found")
	case errors.Is(err, domain.ErrAlreadyExists):
		writeError(w, http.StatusConflict, "ALREADY_EXISTS", "resource already exists")
	case errors.Is(err, domain.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
	default:
		// Неожиданная ошибка: клиенту — нейтральное сообщение,
		// подробности — только в лог (не палим внутренности наружу).
		logger.Error("unexpected error", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL", "internal server error")
	}

}
