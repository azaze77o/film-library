package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/project/omdbapp/internal/domain"
	"github.com/project/omdbapp/internal/service"
)

// Пока нет пользователя, поэтому фиксируем общего
const defaultUserID = domain.UserID("default")

type LibraryHandler struct {
	svc    *service.LibraryService
	logger *slog.Logger
}

type createLibraryEntryRequest struct {
	ImdbID string `json:"imdb_id"`
	Title  string `json:"title"`
	Year   string `json:"year"`
	Type   string `json:"type"`
	Poster string `json:"poster"`
}

type updateLibraryEntryRequest struct {
	Favourite *bool `json:"favourite"`
	Watched   *bool `json:"watched"`
}

func NewLibraryHandler(svc *service.LibraryService, logger *slog.Logger) *LibraryHandler {
	return &LibraryHandler{svc: svc, logger: logger}
}

func (h *LibraryHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	entries, err := h.svc.List(ctx, defaultUserID)

	if err != nil {
		handleServiceError(w, h.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func (h *LibraryHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req createLibraryEntryRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", "invalid json")
		return
	}

	movie := domain.MoviePreview{
		ID:     domain.ImdbID(req.ImdbID),
		Title:  req.Title,
		Year:   req.Year,
		Type:   req.Type,
		Poster: req.Poster,
	}

	entry, err := h.svc.Create(ctx, defaultUserID, movie)
	if err != nil {
		handleServiceError(w, h.logger, err)
		return
	}

	writeJSON(w, http.StatusCreated, entry)
}

func (h *LibraryHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Достаем ID фильма из пути
	rawID := r.PathValue("id")

	entry, err := h.svc.Get(ctx, defaultUserID, domain.ImdbID(rawID))
	if err != nil {
		handleServiceError(w, h.logger, err)
		return
	}

	writeJSON(w, http.StatusOK, entry)
}

func (h *LibraryHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req updateLibraryEntryRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", "invalid json")
		return
	}

	path := domain.LibraryPatch{
		Favourite: req.Favourite,
		Watched:   req.Watched,
	}

	// Достаем ID фильма из пути
	rawID := r.PathValue("id")

	entry, err := h.svc.Update(ctx, defaultUserID, domain.ImdbID(rawID), path)
	if err != nil {
		handleServiceError(w, h.logger, err)
		return
	}

	writeJSON(w, http.StatusOK, entry)
}

func (h *LibraryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Достаем ID фильма из пути
	rawID := r.PathValue("id")

	err := h.svc.Delete(ctx, defaultUserID, domain.ImdbID(rawID))
	if err != nil {
		handleServiceError(w, h.logger, err)
		return
	}

	// Так как у нас нет тела ответа, то writeJSON здесь лучше не звать на успехе:
	// он всегда ставит Content-Type: application/json
	w.WriteHeader(http.StatusNoContent)
}
