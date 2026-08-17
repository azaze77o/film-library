package omdb

import (
	"encoding/csv"
	"os"
)

func ExportMoviesToCSV(filename string, movies []SearchMovie) error {

	// 1. Создаем csv файл

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	// Загловоки столбцов
	if err := writer.Write([]string{
		"title",
		"year",
		"im_db_id",
		"type",
		"poster",
	}); err != nil {
		return err
	}

	// 2. Заполянем файл данными
	for _, movie := range movies {

		// Создаем скоуп данных для записи
		record := []string{
			movie.Title,
			movie.Year,
			movie.ImdbID,
			movie.Type,
			movie.Poster,
		}

		// Записываем данные в буфер
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	// Записываем данные из буфера в файл
	writer.Flush()
	if err := writer.Error(); err != nil {
		return err
	}

	return nil
}
