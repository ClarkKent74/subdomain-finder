package scanner

import (
	"context"

	"subdomain-finder/internal/entity"
)

type Scanner interface {
	Scan(ctx context.Context, domain string) ([]string, error)
}

type Registry map[entity.Algorithm]Scanner
