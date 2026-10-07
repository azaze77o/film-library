package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/project/omdbapp/internal/repository/omdb"
	"github.com/project/omdbapp/internal/repository/postgres"
	"github.com/project/omdbapp/internal/service"
	"github.com/project/omdbapp/internal/transport/httpapi"
)

func main() {
	// Создаем ctx для старта приложения. В запросах будет использоваться свой контекст
	// Создаем logger для логирования работы сервера.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Загружаем ключ API
	if err := godotenv.Load(); err != nil {
		log.Println("Предупреждение файл .env не найден, использую переменные окружения.")
	}
	apiKey := os.Getenv("OMDB_API_KEY")
	if apiKey == "" {
		log.Fatal("OMDB_API_KEY не задан. Создайте файл .env или задайте переменную окружения.")
	}

	// Читаетм адрес внешнего агреатора фильмов OMDb API
	baseURL := os.Getenv("OMDB_BASE_URL")
	if baseURL == "" {
		log.Fatal("OMDB_BASE_URL не задан. Создайте файл .env или задайте URL OMDb API.")
	}

	// Инициализируем подключение к PostgreSQL
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL не задан. Создайте файл .env или задайте адрес подключения.")
	}
	db, err := postgres.New(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	logger.Info("db connected")

	// Собираем зависимости: репозитории -> сервисы
	// Для поиска
	searchClient := omdb.NewClient(baseURL, apiKey)
	searchSrv := service.NewSearchService(searchClient)

	// Для хранения и управления библиотекой фильмов
	store := postgres.NewLibraryStore(db)
	libSrv := service.NewLibraryService(store)

	// Для хранения данных

	// Поднимаем сервер
	router := httpapi.NewRouter(libSrv, searchSrv, logger)

	addr := ":8080"
	logger.Info("starting server", "addr", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}

}
