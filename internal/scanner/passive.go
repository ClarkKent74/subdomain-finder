package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type crtShEntry struct {
	NameValue string `json:"name_value"`
}

type cbState int

const (
	cbClosed cbState = iota
	cbOpen
	cbHalfOpen
)

// circuitBreaker — защита от каскадных отказов внешнего API.
// Если crt.sh возвращает ошибки N раз подряд — перестаём к нему обращаться
// на время timeout, затем пробуем снова.
type circuitBreaker struct {
	mu        sync.Mutex
	state     cbState
	failures  int
	threshold int
	timeout   time.Duration
	openedAt  time.Time
}

// allow возвращает true если запрос можно выполнить.
func (cb *circuitBreaker) allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case cbClosed:
		return true
	case cbOpen:
		// Таймаут истёк — переходим в half-open и пробуем один запрос
		if time.Since(cb.openedAt) >= cb.timeout {
			cb.state = cbHalfOpen
			return true
		}
		return false
	case cbHalfOpen:
		return true
	default:
		return false
	}
}

// success сообщает что запрос прошёл успешно.
func (cb *circuitBreaker) success() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = cbClosed
}

// failure сообщает что запрос завершился ошибкой.
func (cb *circuitBreaker) failure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	if cb.failures >= cb.threshold {
		cb.state = cbOpen
		cb.openedAt = time.Now()
	}
}

// PassiveScanner — пассивный поиск поддоменов через Certificate Transparency логи.
type PassiveScanner struct {
	client *http.Client
	cb     *circuitBreaker
}

// NewPassiveScanner создаёт пассивный сканер с заданными параметрами.
func NewPassiveScanner(timeout time.Duration, cbThreshold int, cbTimeout time.Duration) *PassiveScanner {
	return &PassiveScanner{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
		cb: &circuitBreaker{
			threshold: cbThreshold,
			timeout:   cbTimeout,
		},
	}
}

// Scan ищет поддомены через crt.sh API.
func (s *PassiveScanner) Scan(ctx context.Context, domain string) ([]string, error) {
	if !s.cb.allow() {
		return nil, fmt.Errorf("crt.sh недоступен, повторите позже (circuit breaker открыт)")
	}

	url := fmt.Sprintf("https://crt.sh/?q=%%.%s&output=json", domain)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		s.cb.failure()
		return nil, fmt.Errorf("ошибка запроса к crt.sh: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.cb.failure()
		return nil, fmt.Errorf("crt.sh вернул статус %d", resp.StatusCode)
	}

	var entries []crtShEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		s.cb.failure()
		return nil, fmt.Errorf("ошибка разбора ответа crt.sh: %w", err)
	}

	s.cb.success()
	return deduplicate(extractSubdomains(entries, domain)), nil
}

// extractSubdomains извлекает поддомены из записей crt.sh.
func extractSubdomains(entries []crtShEntry, domain string) []string {
	var result []string
	for _, e := range entries {
		for _, name := range strings.Split(e.NameValue, "\n") {
			name = strings.TrimSpace(name)
			name = strings.TrimPrefix(name, "*.")
			name = strings.ToLower(name)

			if strings.HasSuffix(name, "."+domain) && name != domain {
				result = append(result, name)
			}
		}
	}
	return result
}

// deduplicate убирает дублирующиеся строки из слайса.
func deduplicate(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; !ok {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
