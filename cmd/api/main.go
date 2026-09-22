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
	"github.com/project/omdbapp/transport/httpapi"
)

func main() {
	// 1. Создаем ctx для старта приложения. В запросах будет использоваться свой контекст
	// Создаем logger для логирования работы сервера.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// 2. Загружаем ключ API
	if err := godotenv.Load(); err != nil {
		log.Println("Предупреждение файл .env не найден, использую переменные окружения.")
	}
	apiKey := os.Getenv("OMDB_API_KEY")
	if apiKey == "" {
		log.Fatal("OMDB_API_KEY не задан. Создайте файл .env или задайте переменную окружения.")
	}

	// 3. Инициализируем подключение к PostgreSQL
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

	// 4. Собираем зависимости: репозитории -> сервисы
	// 4.1 Для поиска
	searchClient := omdb.NewClient(apiKey)
	searchSrv := service.NewSearchService(searchClient)

	// 4.2 Для хранения и управления библиотекой фильмов
	store := postgres.NewLibraryStore(db)
	libSrv := service.NewLibraryService(store)

	// 4.3 Для хранения данных

	// 5.Поднимаем сервер
	router := httpapi.NewRouter(libSrv, searchSrv, logger)

	addr := ":8080"
	logger.Info("starting server", "addr", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}

}
