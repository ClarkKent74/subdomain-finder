package api

import (
	"net/http"
	"strings"

	"subdomain-finder/internal/entity"
	"subdomain-finder/internal/service"
)

// максимальный размер загружаемого файла словаря — 10MB
const maxWordlistSize = 10 << 20

// Handler — HTTP обработчики сервиса.
type Handler struct {
	svc *service.FinderService
}

// NewHandler создаёт новый Handler.
func NewHandler(svc *service.FinderService) *Handler {
	return &Handler{svc: svc}
}

// FindDomains godoc
// @Summary      Запустить поиск поддоменов
// @Description  Создаёт задачу на поиск поддоменов. Для алгоритма bruteforce можно загрузить свой словарь.
// @Tags         domains
// @Accept       multipart/form-data
// @Produce      json
// @Param        domain    formData  string  true   "Целевой домен (например: example.com)"
// @Param        algorithm formData  string  true   "Алгоритм: passive, bruteforce, zonetransfer"
// @Param        wordlist  formData  file    false  "Файл словаря для bruteforce (одно слово на строку)"
// @Success      201  {object}  findDomainsResponse
// @Failure      400  {object}  errorResponse
// @Failure      409  {object}  errorResponse  "Задача уже выполняется"
// @Failure      429  {object}  errorResponse  "Очередь переполнена или rate limit"
// @Router       /findDomains [post]
func (h *Handler) FindDomains(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxWordlistSize+1024)

	if err := r.ParseMultipartForm(maxWordlistSize); err != nil {
		writeError(w, http.StatusBadRequest, "ошибка разбора формы: "+err.Error())
		return
	}

	domain := strings.TrimSpace(r.FormValue("domain"))
	algorithmStr := strings.TrimSpace(r.FormValue("algorithm"))

	if domain == "" {
		writeError(w, http.StatusBadRequest, "параметр domain обязателен")
		return
	}
	if algorithmStr == "" {
		writeError(w, http.StatusBadRequest, "параметр algorithm обязателен")
		return
	}

	algorithm := entity.Algorithm(algorithmStr)

	// Читаем файл словаря если передан
	var words []string
	file, _, err := r.FormFile("wordlist")
	if err == nil {
		defer file.Close()
		words, err = parseWordlist(file)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ошибка чтения словаря: "+err.Error())
			return
		}
		if len(words) == 0 {
			writeError(w, http.StatusBadRequest, "файл словаря пустой")
			return
		}
	}

	if err := h.svc.FindDomains(r.Context(), domain, algorithm, words); err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, findDomainsResponse{
		Domain:    domain,
		Algorithm: algorithmStr,
		Status:    string(entity.StatusPending),
	})
}

// GetResult godoc
// @Summary      Получить результат сканирования
// @Description  Возвращает статус и результаты задачи. 202 — ещё выполняется, 200 — готово.
// @Tags         domains
// @Produce      json
// @Param        domain    query  string  true  "Целевой домен"
// @Param        algorithm query  string  true  "Алгоритм: passive, bruteforce, zonetransfer"
// @Success      200  {object}  getResultResponse  "Сканирование завершено"
// @Success      202  {object}  getResultResponse  "Сканирование ещё выполняется"
// @Failure      400  {object}  errorResponse
// @Failure      404  {object}  errorResponse  "Задача не найдена"
// @Router       /getResult [get]
func (h *Handler) GetResult(w http.ResponseWriter, r *http.Request) {
	domain := strings.TrimSpace(r.URL.Query().Get("domain"))
	algorithmStr := strings.TrimSpace(r.URL.Query().Get("algorithm"))

	if domain == "" {
		writeError(w, http.StatusBadRequest, "параметр domain обязателен")
		return
	}
	if algorithmStr == "" {
		writeError(w, http.StatusBadRequest, "параметр algorithm обязателен")
		return
	}

	task, err := h.svc.GetResult(r.Context(), domain, entity.Algorithm(algorithmStr))
	if err != nil {
		writeServiceError(w, err)
		return
	}

	resp := getResultResponse{
		Domain:    task.Domain,
		Algorithm: string(task.Algorithm),
		Status:    string(task.Status),
		Results:   task.Results,
		Error:     task.Error,
	}

	statusCode := http.StatusOK
	if task.Status.IsActive() {
		statusCode = http.StatusAccepted
	}

	writeJSON(w, statusCode, resp)
}
