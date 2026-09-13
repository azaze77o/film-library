package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/project/omdbapp/internal/repository/memory"
	"github.com/project/omdbapp/internal/repository/omdb"
	"github.com/project/omdbapp/internal/service"
	"github.com/project/omdbapp/transport/httpapi"
)

func main() {
	// 1. Загружаем ключ API
	if err := godotenv.Load(); err != nil {
		log.Println("Предупреждение файл .env не найден, использую переменные окружения.")
	}
	apiKey := os.Getenv("OMDB_API_KEY")
	if apiKey == "" {
		log.Fatal("OMDB_API_KEY не задан. Создайте файл .env или задайте переменную окружения.")
	}

	// 2. Собираем зависимости: репозитории -> сервисы
	// 2.1 Для поиска
	searchClient := omdb.NewClient(apiKey)
	searchSrv := service.NewSearchService(searchClient)

	// 2.2 Для Управления библиотекой фильмов
	store := memory.NewLibraryStore()
	libSrv := service.NewLibraryService(store)

	// 2.3 Для логирования работы сервера
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	router := httpapi.NewRouter(libSrv, searchSrv, logger)

	addr := ":8080"
	logger.Info("stratring server", "addr", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}

}
