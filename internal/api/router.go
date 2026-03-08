package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// NewRouter создаёт и настраивает HTTP роутер.
func NewRouter(h *Handler, requestsPerMinute int) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(LoggingMiddleware)
	r.Use(RateLimitMiddleware(requestsPerMinute))

	r.Post("/findDomains", h.FindDomains)
	r.Get("/getResult", h.GetResult)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	return r
}
