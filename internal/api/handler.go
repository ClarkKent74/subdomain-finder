package api

import (
	"net/http"
	"strings"

	"subdomain-finder/internal/entity"
	"subdomain-finder/internal/service"
)

type Handler struct {
	svc *service.FinderService
}

func NewHandler(svc *service.FinderService) *Handler {
	return &Handler{svc: svc}
}

// FindDomains godoc
// @Summary      Запустить поиск поддоменов
// @Description  Создаёт задачу на поиск поддоменов через Sudomy.
// @Tags         domains
// @Produce      json
// @Param        domain query string true "Целевой домен (например: example.com)"
// @Success      201  {object}  findDomainsResponse
// @Failure      400  {object}  errorResponse
// @Failure      409  {object}  errorResponse
// @Failure      429  {object}  errorResponse
// @Router       /findDomains [post]
func (h *Handler) FindDomains(w http.ResponseWriter, r *http.Request) {
	domain := strings.TrimSpace(r.URL.Query().Get("domain"))
	if domain == "" {
		writeError(w, http.StatusBadRequest, "параметр domain обязателен")
		return
	}

	if err := h.svc.FindDomains(r.Context(), domain); err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, findDomainsResponse{
		Domain: domain,
		Status: string(entity.StatusPending),
	})
}

// GetResult godoc
// @Summary      Получить результат сканирования
// @Description  Возвращает статус и результаты задачи. 202 — ещё выполняется, 200 — готово.
// @Tags         domains
// @Produce      json
// @Param        domain query string true "Целевой домен"
// @Success      200  {object}  getResultResponse
// @Success      202  {object}  getResultResponse
// @Failure      400  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Router       /getResult [get]
func (h *Handler) GetResult(w http.ResponseWriter, r *http.Request) {
	domain := strings.TrimSpace(r.URL.Query().Get("domain"))
	if domain == "" {
		writeError(w, http.StatusBadRequest, "параметр domain обязателен")
		return
	}

	task, err := h.svc.GetResult(r.Context(), domain)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	resp := getResultResponse{
		Domain:  task.Domain,
		Status:  string(task.Status),
		Results: task.Results,
		Error:   task.Error,
	}

	statusCode := http.StatusOK
	if task.Status.IsActive() {
		statusCode = http.StatusAccepted
	}

	writeJSON(w, statusCode, resp)
}
