package csv

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/project/omdbapp/internal/domain"
)

func ExportLibraryToCSV(filename string, entries []domain.LibraryEntry) error {
	// 1. Создаём CSV-файл
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	// Закрыть файл при выходе из функции.
	defer file.Close()

	// 2. Создаём писателя данных в буфер памяти
	writer := csv.NewWriter(file)

	// 3. Делаем отложенный вывод данных из буфера в файл
	defer writer.Flush()

	// Заголовки столбцов
	if err := writer.Write([]string{"im_db_id", "title", "year", "favorite", "watched"}); err != nil {
		return err
	}

	// 4. Заполняем файл данными
	for _, e := range entries {

		// Создаём срез данных для записи
		record := []string{
			string(e.MoviePreview.ID),
			e.MoviePreview.Title,
			e.MoviePreview.Year,
			fmt.Sprintf("%t", e.IsFavorite),
			fmt.Sprintf("%t", e.IsWatched),
		}

		// Записываем данные в буфер
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return writer.Error()
}
