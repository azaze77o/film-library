package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/project/omdbapp/internal/queue"
	"github.com/project/omdbapp/internal/repository/omdb"
	"github.com/project/omdbapp/internal/repository/postgres"
	"github.com/project/omdbapp/internal/service"
	"github.com/project/omdbapp/internal/transport/httpapi"
	"github.com/project/omdbapp/internal/worker"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(logger); err != nil {
		logger.Error("app stopped with error", "error", err)
		os.Exit(1)
	}
	logger.Info("app stopped")
}

func run(logger *slog.Logger) error {
	// ctx отменится, когда придёт SIGINT (Ctrl+C) или SIGTERM
	// Сервер начнет завершение текущих процессов и выход из run
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Загружаем ключ API
	if err := godotenv.Load(); err != nil {
		log.Println("Предупреждение файл .env не найден, использую переменные окружения.")
	}
	apiKey := os.Getenv("OMDB_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OMDB_API_KEY не задан. Создайте файл .env или задайте переменную окружения.")
	}

	// Читаетм адрес внешнего агреатора фильмов OMDb API
	baseURL := os.Getenv("OMDB_BASE_URL")
	if baseURL == "" {
		return fmt.Errorf("OMDB_BASE_URL не задан. Создайте файл .env или задайте URL OMDb API.")
	}

	// Инициализируем подключение к PostgreSQL
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL не задан. Создайте файл .env или задайте адрес подключения.")
	}

	// Для подключения к БД инициализируем свой дочерний контектс с таймаутом
	// startCtx отменится в двух случаях:
	// - либо прошли пять секунд
	// - либо отменился родитель, то есть пришёл Ctrl+C во время проверки.
	startCtx, cancelStart := context.WithTimeout(ctx, 5*time.Second)
	defer cancelStart()
	db, err := postgres.New(startCtx, dbURL)
	if err != nil {
		return fmt.Errorf("connecting database: %w", err)
	}
	defer db.Close()
	logger.Info("db connected")

	// Собираем зависимости: репозитории -> сервисы
	// Для поиска
	searchClient := omdb.NewClient(baseURL, apiKey)
	searchSrv := service.NewSearchService(searchClient)

	store := postgres.NewLibraryStore(db) // Создание хранилища
	events := queue.New(100)              // Создание очереди

	auditStore := postgres.NewAuditStore(db)
	auditWorker := worker.NewAuditWorker(
		events.Events(),
		auditStore,
		logger,
		3*time.Second,
	)

	// Инициализируем отдельный контекст для воркера
	// Иначе при нажатии Ctrl+C воркер немедленно получит отмену и бросит всё,
	// что осталось в буфере,
	// а нам нужно, чтобы он дочитал буфер после закрытия очереди
	workerCtx, stopWorker := context.WithCancel(context.Background())
	defer stopWorker()

	// Запуск воркера в отдельной горутине
	var wg sync.WaitGroup
	wg.Go(func() {
		auditWorker.Run(workerCtx)
	})

	// Поднятие сервиса управления библиотекой
	libSrv := service.NewLibraryService(store, events, logger)

	// Инициализируем роутер
	router := httpapi.NewRouter(libSrv, searchSrv, logger)

	// Поднимаем сервер
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,  // заголовки запроса
		ReadTimeout:       10 * time.Second, // весь запрос вместе с телом
		WriteTimeout:      15 * time.Second, // от конца чтения запроса до конца ответа
		IdleTimeout:       60 * time.Second, // keep-alive между запросами
	}

	// Канал хранит одну ошибку сервера.
	// Буфер 1 позволяет горутине отправить результат,
	// даже если run к этому моменту уже перешёл к остановке.
	serverErr := make(chan error, 1)

	go func() {
		logger.Info("starting server", "addr", srv.Addr)
		serverErr <- srv.ListenAndServe()
	}()

	var runErr error

	select {
	case err := <-serverErr:
		runErr = fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	// Перед ожиданием остановки сервреа прекращаем обработку сигналов.
	stop()

	// Создаем независимый контекст для отдельного времени на заверешение работы сервера
	// Нельзя использовать родительский, так как к этому времени он может быть уже отменен (выше)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("http shutdown", "error", err)
	}

	// Пишем в лог, что сервер остановлен
	logger.Info("http server stopped")

	// Закрываем очередь и пишем об этом в лог
	logger.Info("closing event queue")
	events.Close()

	// По сигналу workerDone ожидаем заверешения обработки очереди воркером
	workerDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(workerDone)
	}()

	select {
	case <-workerDone:
		logger.Info("event queue drained") // воркер дочитал очередь и вышел сам
	case <-shutdownCtx.Done():
		logger.Warn("worker did not finish in time, cancelling") // закончился бюджет на закрытие
		stopWorker()
		<-workerDone // дождаться выхода, иначе db.Close() сработает под воркером
		logger.Warn("worker cancelled, some events may be lost")
	}

	return runErr
}
