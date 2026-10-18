// Сервис поиска фильмов, реализует сценарии:
// 1. Поиск по названию фильмов с перебором страниц с результатами
package service

import (
	"context"
	"strings"

	"github.com/project/omdbapp/internal/domain"
)

type SearchRepository interface {
	Search(ctx context.Context, title string, page int) (domain.SearchResult, error)
}

type SearchService struct {
	repo SearchRepository
}

func NewSearchService(repo SearchRepository) *SearchService {
	return &SearchService{repo: repo}
}

func (s *SearchService) Search(ctx context.Context, title string, page int) (domain.SearchResult, error) {
	if strings.TrimSpace(title) == "" {
		return domain.SearchResult{}, domain.ErrInvalidInput
	}

	return s.repo.Search(ctx, title, page)
}
