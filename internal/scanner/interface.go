package scanner

import (
	"context"
)

type Scanner interface {
	Scan(ctx context.Context, domain string) ([]string, error)
}
