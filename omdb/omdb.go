/*Пакет является клиентом, который делает запрос в OMDbAPI по названию фильма
и вовзвращает данные фильма, которые соппадает по названию*/

package omdb

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const baseURL = "http://www.omdbapi.com/"

var (
	apiKey string
	client *http.Client
)

type SearchMovie struct {
	Title  string `json:"Title"`
	Year   string `json:"Year"`
	ImdbID string `json:"imdbID"`
	Type   string `json:"Type"`
	Poster string `json:"Poster"`
}

type MovieDetails struct {
	Title    string   `json:"Title"`
	Year     string   `json:"Year"`
	Genre    string   `json:"Genre"`
	Director string   `json:"Director"`
	Actors   string   `json:"Actors"`
	Plot     string   `json:"Plot"`
	Country  string   `json:"Country"`
	Ratings  []Rating `json:"Ratings"`
	Response string   `json:"Response"`
	Error    string   `json:"Error"`
}

type Rating struct {
	Source string `json:"Source"`
	Value  string `json:"Value"`
}

type SearchResponse struct {
	Search       []SearchMovie `json:"Search"`
	TotalResults string        `json:"totalResults"`
	Response     string        `json:"Response"`
	Error        string        `json:"Error"`
}

// Функция для инициализации ключа API
func Init() error {

	// Читаем файл .env с обработкой ошибки, если его нет
	if err := godotenv.Load(); err != nil {
		log.Println("Предупреждение: файл .env не найден, пытаюсь использовать переменные среды операционной системы.")
	}

	// Из файла читаем нужную перменную, в котоой хранится ключ с обработкой ошибки, если ключ не найден
	apiKey = os.Getenv("OMDB_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OMDB_API_KEY не задан. Создайте файл .env или задайте переменную окружения.")
	}

	client = &http.Client{}
	return nil
}

// GetMovieByTitle возвращает фильм по названию
// Дополнительно используются параметры для фильтрации
// Year  - год выпуска
// Plot - размера описанию сюжета (Full - полное описание, Short - короткое)

func SearchMovies(title string, page int) (SearchResponse, error) {
	if client == nil {
		return SearchResponse{}, fmt.Errorf("Вызовите Init() перед использованием")
	}

	if page < 1 {
		page = 1
	}

	params := url.Values{}
	params.Add("apiKey", apiKey)
	params.Add("s", title)
	params.Add("type", "movie")
	params.Add("page", strconv.Itoa(page))

	// Инициализхируем переменную для создания URL запроса
	// params.Encode() кодирует все ключ-значения параметров для использования в URL
	// внутри уже разделяя каждый параметр через &
	u := baseURL + "?" + params.Encode()

	//Делаем запрос через клиента, в параметр передаем строку URL
	resp, err := client.Get(u)
	if err != nil {
		return SearchResponse{}, fmt.Errorf("Ошибка запроса: %w", err)
	}

	// Закрываем соединеие
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return SearchResponse{}, fmt.Errorf("OMDb вернул HTTP-статус: %s", resp.Status)
	}

	var result SearchResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return SearchResponse{}, fmt.Errorf("Ошибка парсинга JSON: %w", err)
	}

	if result.Response == "False" {
		return SearchResponse{}, fmt.Errorf("Фильмы не найдены: %s", result.Error)
	}

	return result, nil
}
