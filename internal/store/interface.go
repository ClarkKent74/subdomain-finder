package store

import (
	"context"

	"subdomain-finder/internal/entity"
)

type Store interface {
	Create(ctx context.Context, task *entity.Task) error
	Get(ctx context.Context, domain string, algorithm entity.Algorithm) (*entity.Task, error)
	Update(ctx context.Context, task *entity.Task) error
}
