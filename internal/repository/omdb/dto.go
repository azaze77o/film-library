package omdb

import (
	"strconv"

	"github.com/project/omdbapp/internal/domain"
)

type moviePreviewDTO struct {
	ID     string `json:"imdbID"`
	Title  string `json:"Title"`
	Year   string `json:"Year"`
	Type   string `json:"Type"`
	Poster string `json:"Poster"`
}

type searchResponseDTO struct {
	Search       []moviePreviewDTO `json:"Search"`
	TotalResults string            `json:"totalResults"`
	Response     string            `json:"Response"`
	Error        string            `json:"Error"`
}

type movieDTO struct {
	ID       string      `json:"imdbID"`
	Title    string      `json:"Title"`
	Year     string      `json:"Year"`
	Genre    string      `json:"Genre"`
	Director string      `json:"Director"`
	Actors   string      `json:"Actors"`
	Plot     string      `json:"Plot"`
	Country  string      `json:"Country"`
	Ratings  []ratingDTO `json:"Ratings"`
	Response string      `json:"Response"`
	Error    string      `json:"Error"`
}

type ratingDTO struct {
	Source string `json:"Source"`
	Value  string `json:"Value"`
}

// toDomain - метод приведения превью внешней модели данных к доменной модели сервиса
func (m *moviePreviewDTO) toDomain() domain.MoviePreview {
	return domain.MoviePreview{
		ID:     domain.ImdbID(m.ID),
		Title:  m.Title,
		Type:   m.Type,
		Poster: m.Poster,
	}

}

// toDomain - метод приведения ответа поиска внешней модели данных к доменной модели сервиса
func (s *searchResponseDTO) toDomain() (domain.SearchResult, error) {
	movies := make([]domain.MoviePreview, 0, len(s.Search))
	for _, m := range s.Search {
		movies = append(movies, m.toDomain())
	}

	// Приводим количество найденных фильмов в целочисленный формат
	total, err := strconv.Atoi(s.TotalResults)
	if err != nil {
		total = 0
	}

	return domain.SearchResult{
		Movies:       movies,
		TotalResults: total,
	}, nil
}

// toDomain - метод приведения карточки фильма внешней модели данных к доменной модели сервиса
func (m *movieDTO) toDomain() domain.Movie {
	ratings := make([]domain.Rating, 0, len(m.Ratings))
	for _, r := range m.Ratings {
		ratings = append(ratings, domain.Rating{
			Source: r.Source,
			Value:  r.Value,
		})
	}

	return domain.Movie{
		ID:       domain.ImdbID(m.ID),
		Title:    m.Title,
		Year:     m.Year,
		Genre:    m.Genre,
		Director: m.Director,
		Actors:   m.Actors,
		Plot:     m.Plot,
		Country:  m.Country,
		Ratings:  ratings,
	}

}
