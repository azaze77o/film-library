package httpapi

import (
	"encoding/json"
	"net/http"
)

// writeJSON - общая функция, возвращающая ответ клиенту
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body != nil {
		_ = json.NewEncoder(w).Encode(body)
	}
}

// errorBody - общая структура ошибки, возвращаемая клиенту в теле ответа
type errorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// writeError - функция, пишущая ошибку на запрос.
// Использует структуру errorBody для ответа и функцию writeJSON для отправки
func writeError(w http.ResponseWriter, status int, code, message string) {
	var b errorBody
	b.Error.Code = code
	b.Error.Message = message
	writeJSON(w, status, b)
}
