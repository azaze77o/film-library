/*
Программа для отображения результатов поиска фильма по его названию
C помощью API сайта https://www.omdbapi.com/
Пример запуска: go run main.go "The Matrix" 1
*/

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/project/omdbapp/omdb"
)

func main() {

	// 1. Инициализируем клиент и ключ
	if err := omdb.Init(); err != nil {
		log.Fatal(err)
	}

	// 2. Проверяем, что пользователь что‑то ввёл
	if len(os.Args) < 3 {
		fmt.Println("Использование: go run main.go \"название фильма\" <страница>")
		return
	}

	// 3. Собираем аругменты для поиска
	title := os.Args[1]
	page, err := strconv.Atoi(os.Args[2])

	if err != nil {
		fmt.Println("Номер страницы должен быть числом")
		return
	}

	result, err := omdb.SearchMovies(title, page)

	if err != nil {
		log.Fatal(err)
	}

	// Вывод результатов
	fmt.Printf("Страница %d. Найдено фильмов: %s\n", page, result.TotalResults)
	if len(result.Search) == 0 {
		fmt.Println("Результатов нет")
	} else {
		for _, m := range result.Search {
			fmt.Printf("- %s (%s)\n", m.Title, m.Year)
		}
	}

	// Экспортируем результаты в файл
	err = omdb.ExportMoviesToCSV("movies.csv", result.Search)
	if err != nil {
		fmt.Println("Ошибка экспорта:", err)
		return
	}

	fmt.Println("Создан CSV-файл")

}
