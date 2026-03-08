package validate

import (
	"fmt"
	"regexp"
	"strings"
)

var domainRegex = regexp.MustCompile(
	`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`,
)

func Domain(domain string) error {
	domain = strings.TrimSpace(domain)

	if domain == "" {
		return fmt.Errorf("домен не может быть пустым")
	}

	if len(domain) > 253 {
		return fmt.Errorf("домен слишком длинный: %d символов, максимум 253", len(domain))
	}

	for _, label := range strings.Split(domain, ".") {
		if len(label) > 63 {
			return fmt.Errorf("метка домена слишком длинная: %q, максимум 63 символа", label)
		}
	}

	if !domainRegex.MatchString(domain) {
		return fmt.Errorf("некорректный формат домена: %q", domain)
	}

	return nil
}
