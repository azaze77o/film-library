package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/project/omdbapp/internal/service"
)

func NewRouter(lib *service.LibraryService, search *service.SearchService, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	// Обработчики сервиса по управлению библиотекой пользователя
	l := NewLibraryHandler(lib, logger)

	mux.HandleFunc("GET /api/v1/library", l.List)
	mux.HandleFunc("POST /api/v1/library", l.Create)
	mux.HandleFunc("GET /api/v1/library/{id}", l.Get)
	mux.HandleFunc("PATCH /api/v1/library/{id}", l.Update)
	mux.HandleFunc("DELETE /api/v1/library/{id}", l.Delete)

	// Обработчики сервиса поиска фильмов
	s := NewSearchHandler(search, logger)
	mux.HandleFunc("GET /api/v1/search", s.Search)

	return withLogging(mux, logger)
}
