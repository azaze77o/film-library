package omdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/project/omdbapp/internal/domain"
)

const baseURL = "http://www.omdbapi.com/"

// *http.Client. Это «отправитель» — он умеет реально передавать запросы по сети и получать ответы, управлять соединениями
type Client struct {
	apiKey string
	http   *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Search(title string, page int) (domain.SearchResult, error) {
	if page < 1 {
		page = 1
	}

	params := url.Values{}
	params.Add("apiKey", c.apiKey)
	params.Add("s", title)
	params.Add("type", "movie")
	params.Add("page", strconv.Itoa(page))

	// Инициализируем переменную для создания URL запроса
	// params.Encode() кодирует все ключ-значения параметров для использования в URL
	// внутри уже разделяя каждый параметр через &
	u := baseURL + "?" + params.Encode()

	// 1. Формируем GET запрос, по собранному URL - u, но не отправляем его.
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return domain.SearchResult{}, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// 2. Делаем запрос, получаем ответ
	resp, err := c.http.Do(req)
	if err != nil {
		return domain.SearchResult{}, fmt.Errorf("Ошибка запроса: %w", err)
	}

	// Закрываем соединение
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.SearchResult{}, fmt.Errorf("OMDb вернул HTTP-статус: %s", resp.Status)
	}

	// Объявляем переменную для записи декодированного ответа
	var dto searchResponseDTO

	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return domain.SearchResult{}, fmt.Errorf("Ошибка парсинга JSON: %w", err)
	}

	if dto.Response == "False" {
		return domain.SearchResult{}, fmt.Errorf("Фильмы не найдены: %s", dto.Error)
	}

	return dto.toDomain()
}
