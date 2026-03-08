package api

import (
	"bufio"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"subdomain-finder/internal/entity"
)

// parseWordlist читает файл словаря и возвращает список слов.
func parseWordlist(r interface{ Read([]byte) (int, error) }) ([]string, error) {
	var words []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word == "" || strings.HasPrefix(word, "#") {
			continue
		}
		words = append(words, word)
	}
	return words, scanner.Err()
}

// writeServiceError переводит ошибки сервиса в HTTP статусы.
func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, entity.ErrTaskNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, entity.ErrTaskAlreadyExists):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrQueueFull), errors.Is(err, entity.ErrStoreFull):
		writeError(w, http.StatusTooManyRequests, err.Error())
	case errors.Is(err, entity.ErrInvalidDomain), errors.Is(err, entity.ErrInvalidAlgorithm):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
	}
}

// writeJSON сериализует v в JSON и отправляет ответ.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError отправляет ошибку в формате JSON.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}
