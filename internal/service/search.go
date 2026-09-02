// Сервис поиска фильмов, реализует сценарии:
// 1. Поиск по названию фильмов с перебором страниц с результатами
package service

import (
	"github.com/project/omdbapp/internal/domain"
)

type SearchRepository interface {
	Search(title string, page int) (domain.SearchResult, error)
}

type SearchService struct {
	repo SearchRepository
}

func NewSearchService(repo SearchRepository) *SearchService {
	return &SearchService{repo: repo}
}

func (s *SearchService) Search(title string, page int) (domain.SearchResult, error) {
	return s.repo.Search(title, page)
}
