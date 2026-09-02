package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/project/omdbapp/internal/domain"
	"github.com/project/omdbapp/internal/repository/csv"
	"github.com/project/omdbapp/internal/repository/memory"
	"github.com/project/omdbapp/internal/repository/omdb"
	"github.com/project/omdbapp/internal/service"
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
	searchClient := omdb.NewClient(apiKey)
	searchSrv := service.NewSearchService(searchClient)

	libStore := memory.NewLibraryStore()
	libSrv := service.NewLibraryService(libStore)

	// 3. Объявляем фиксированного пользователя
	user := domain.User{ID: domain.UserID("default"), Name: "me"}

	// 4. REPL-цикл
	// Результаты последнего поиска (по ним адресуются команды fav/watch)
	var lastResult []domain.MoviePreview

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Film Library. Введите help для списка команд.")

	// Бесконечный цикл, который обрабатывает команды пользователя выполнением нужного сервиса.
	for {
		fmt.Print(">")
		// Читаем строку, если ввод закончился, выходим из цикла
		if !scanner.Scan() {
			break
		}

		// Убираем пробелы по краям и проверяем наличие ввода, в противном начинаем следующую итерацию цикла (т.е. сначала)
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Разделяем ввод на слова, ожидаем, что команда - это первая часть ввода
		parts := strings.Fields(input)
		cmd := parts[0]

		// В зависимости от введенной команды вызываем нужный сервис
		switch cmd {
		// Выполнение поиска фильма по названию
		case "search":
			if len(parts) < 2 {
				fmt.Println("Использование: search <название> <страница>")
				continue
			}

			title := parts[1]
			page := 1 // Указываем значение по умолчанию, если пользователь не указал номер страницы

			// Определяем номер страницы, если пользователь ее указал
			if len(parts) >= 3 {
				if p, err := strconv.Atoi(parts[2]); err == nil && p > 0 {
					page = p
				}
			}

			// Выполняем поиск
			result, err := searchSrv.Search(title, page)
			if err != nil {
				fmt.Println("Ошибка:", err)
				continue
			}

			// Перезаписываем последние результаты поиска
			lastResult = result.Movies
			fmt.Printf("Страница %d. Найдено фильмов: %d\n", page, result.TotalResults)
			if len(lastResult) == 0 {
				fmt.Println("Результатов нет")
			}

			// Печатаем результат поиска на экран
			for i, m := range lastResult {
				fmt.Printf("  %d. %s (%s) [%s]\n", i+1, m.Title, m.Year, m.ID)
			}

		// Выполнение команды добавления фильма в избранное
		case "fav":
			if len(parts) != 2 {
				fmt.Println("Использование: fav <imdbID>")
				continue
			}
			// Ищем ID фильма, чтобы отметить его как избранный
			m, ok := pickResult(parts[1], lastResult)
			if !ok {
				continue
			}

			if err := libSrv.AddFavorite(user.ID, m); err != nil {
				fmt.Println("Ошибка", err)
			}
			fmt.Printf("%s отмечен как избранный\n", m.Title)

		// Выполнение конмады отметки о том, что фильм просмотрен
		case "watch":
			if len(parts) != 2 {
				fmt.Println("Использование: watch <imdbID>")
				continue
			}
			// Ищем ID фильма, чтобы  отметить его как просмотренный
			m, ok := pickResult(parts[1], lastResult)
			if !ok {
				continue
			}

			if err := libSrv.MarkWatched(user.ID, m); err != nil {
				fmt.Println("Ошибка", err)
			}
			fmt.Printf("%s отмечен как просмотренный\n", m.Title)

		// Выполнение команды просмотра библиотеки пользователя
		case "lib":
			// Получаем все фильмы из библиотеки пользователя
			entries, err := libSrv.ListMovies(user.ID)
			if err != nil {
				fmt.Println("Ошибка", err)
				continue
			}

			if len(entries) == 0 {
				fmt.Println("Библиотека пуста")
			}

			// Если библиотека не пуста, то выводим ее на экран
			for _, e := range entries {
				fmt.Printf("%s %s (%s), избранное: %t, просмотренное: %t\n", e.MoviePreview.ID, e.MoviePreview.Title, e.MoviePreview.Year, e.IsFavorite, e.IsWatched)
			}

		// Выполнение экспорта библиотеки пользователя в CSV-файл
		case "export":
			entries, err := libSrv.ListMovies(user.ID)
			if err != nil {
				fmt.Println("Ошибка", err)
				continue
			}

			if len(entries) == 0 {
				fmt.Println("Библиотека пуста")
				continue
			}

			filename := "movies.csv"

			if err := csv.ExportLibraryToCSV(filename, entries); err != nil {
				fmt.Println("Ошибка экспорта", err)
				continue
			}
			fmt.Println("Создан CSV-файл")

		// Выполнение команды для получения справки
		case "help":
			fmt.Println("Доступные команды:")
			fmt.Printf("  %-30s %s\n", "search <название> [страница]", "— поиск фильмов")
			fmt.Printf("  %-30s %s\n", "fav <imdbID>", "— добавить в избранное по ID")
			fmt.Printf("  %-30s %s\n", "watch <imdbID>", "— отметить просмотренным по ID")
			fmt.Printf("  %-30s %s\n", "lib", "— показать библиотеку")
			fmt.Printf("  %-30s %s\n", "export", "— экспорт результатов в CSV")
			fmt.Println("  help | exit")

		case "exit":
			fmt.Println("Пока!")
			return

		default:
			fmt.Printf("Неизвестная команда: %s. Введите help\n", cmd)
		}
	}

	// Проверяем, не вышли ли мы из цикла из-за ошибки чтения (а не из-за EOF)
	if err := scanner.Err(); err != nil {
		fmt.Println("Ошибка чтения ввода:", err)
	}
}

// pickResult ищет фильм в результатах последнего поиска по переданному imdbID
func pickResult(id string, results []domain.MoviePreview) (domain.MoviePreview, bool) {
	for _, m := range results {
		if string(m.ID) == id {
			return m, true
		}
	}
	fmt.Printf("Фильм с ID %s не найден в результатах последнего поиска. Сначала выполните search.\n", id)
	return domain.MoviePreview{}, false
}
