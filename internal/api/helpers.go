package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"subdomain-finder/internal/entity"
)

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
