package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/httprate"
)

// RateLimitMiddleware ограничивает количество запросов с одного IP.
// При превышении лимита возвращает 429 Too Many Requests.
func RateLimitMiddleware(requestsPerMinute int) func(http.Handler) http.Handler {
	return httprate.LimitByIP(requestsPerMinute, time.Minute)
}

// LoggingMiddleware логирует каждый входящий запрос.
// Записывает метод, путь, статус и время выполнения.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Оборачиваем ResponseWriter чтобы перехватить статус код
		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		slog.Info("запрос обработан",
			"метод", r.Method,
			"путь", r.URL.Path,
			"статус", wrapped.status,
			"время", time.Since(start),
			"ip", r.RemoteAddr,
		)
	})
}

// responseWriter — обёртка над http.ResponseWriter для перехвата статус кода.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}
