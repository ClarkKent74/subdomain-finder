package api

// errorResponse — тело ответа при ошибке.
type errorResponse struct {
	Error string `json:"error"`
}

// findDomainsResponse — тело ответа при успешном создании задачи.
type findDomainsResponse struct {
	Domain    string `json:"domain"`
	Algorithm string `json:"algorithm"`
	Status    string `json:"status"`
}

// getResultResponse — тело ответа при получении результата.
type getResultResponse struct {
	Domain    string   `json:"domain"`
	Algorithm string   `json:"algorithm"`
	Status    string   `json:"status"`
	Results   []string `json:"results,omitempty"`
	Error     string   `json:"error,omitempty"`
}
