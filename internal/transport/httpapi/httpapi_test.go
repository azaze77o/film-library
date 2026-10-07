package httpapi

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/project/omdbapp/internal/domain"
	"github.com/stretchr/testify/assert"
)

// TestHandleServiceError - проверяет мапинг доменных ошибок на http статус
func TestHandleServiceError(t *testing.T) {
	// Создаем слайс структур, где каждая структура - это тестовый сценарий
	tests := []struct {
		name       string
		err        error  // что вернул бы сервис
		wantStatus int    // http.StatusNotFound и т.п.
		wantBody   string // тело ответа целиком, JSON в обратных кавычках
		wantLogged bool   // true: текст err должен попасть в лог
	}{
		{
			name:       "not found",
			err:        domain.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   `{"error":{"code":"NOT_FOUND","message":"resource not found"}}`,
		},
		{
			name:       "already exists",
			err:        domain.ErrAlreadyExists,
			wantStatus: http.StatusConflict,
			wantBody:   `{"error":{"code":"ALREADY_EXISTS","message":"resource already exists"}}`,
		},
		{
			name:       "invalid input",
			err:        domain.ErrInvalidInput,
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"INVALID_INPUT","message":"invalid input"}}`,
		},
		{
			name:       "unknown",
			err:        errors.New("connection refused"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":{"code":"INTERNAL","message":"internal server error"}}`,
			wantLogged: true,
		},
		{
			name:       "wrapped not found",
			err:        fmt.Errorf("get library entry: %w", domain.ErrNotFound),
			wantStatus: http.StatusNotFound,
			wantBody:   `{"error":{"code":"NOT_FOUND","message":"resource not found"}}`,
		},
		{
			name:       "invalid input with reason",
			err:        fmt.Errorf("%w: invalid object ID", domain.ErrInvalidInput),
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"INVALID_INPUT","message":"invalid input: invalid object ID"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, nil))
			rec := httptest.NewRecorder()

			handleServiceError(rec, logger, tt.err)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, tt.wantBody, rec.Body.String())

			if tt.wantLogged {
				assert.Contains(t, buf.String(), tt.err.Error())
			} else {
				assert.Empty(t, buf.String())
			}

		})
	}
}
