package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"subdomain-finder/internal/entity"
)

type RedisStore struct {
	client    *redis.Client
	taskTTL   time.Duration
	activeTTL time.Duration
	maxTasks  int
}

func NewRedisStore(ctx context.Context, redisURL string, taskTTL, activeTTL time.Duration, maxTasks int) (*RedisStore, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("некорректный Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("не удалось подключиться к Redis: %w", err)
	}

	return &RedisStore{
		client:    client,
		taskTTL:   taskTTL,
		activeTTL: activeTTL,
		maxTasks:  maxTasks,
	}, nil
}

func (s *RedisStore) Close() error {
	return s.client.Close()
}

func taskKey(domain string, algorithm entity.Algorithm) string {
	return fmt.Sprintf("task:%s:%s", domain, algorithm)
}

const counterKey = "tasks:count"

func (s *RedisStore) Create(ctx context.Context, task *entity.Task) error {
	key := taskKey(task.Domain, task.Algorithm)

	existing, err := s.getByKey(ctx, key)
	if err != nil && !errors.Is(err, entity.ErrTaskNotFound) {
		return fmt.Errorf("ошибка проверки существующей задачи: %w", err)
	}
	if existing != nil && existing.Status.IsActive() {
		return entity.ErrTaskAlreadyExists
	}

	count, err := s.client.Get(ctx, counterKey).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("ошибка чтения счётчика задач: %w", err)
	}
	if count >= s.maxTasks {
		return entity.ErrStoreFull
	}

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("ошибка сериализации задачи: %w", err)
	}

	if err := s.client.Set(ctx, key, data, s.activeTTL).Err(); err != nil {
		return fmt.Errorf("ошибка сохранения задачи в Redis: %w", err)
	}

	s.client.Incr(ctx, counterKey)

	return nil
}

func (s *RedisStore) Get(ctx context.Context, domain string, algorithm entity.Algorithm) (*entity.Task, error) {
	return s.getByKey(ctx, taskKey(domain, algorithm))
}

func (s *RedisStore) Update(ctx context.Context, task *entity.Task) error {
	key := taskKey(task.Domain, task.Algorithm)

	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("ошибка проверки задачи: %w", err)
	}
	if exists == 0 {
		return entity.ErrTaskNotFound
	}

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("ошибка сериализации задачи: %w", err)
	}

	ttl := s.activeTTL
	if !task.Status.IsActive() {
		ttl = s.taskTTL
	}

	if err := s.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("ошибка обновления задачи в Redis: %w", err)
	}

	return nil
}

func (s *RedisStore) getByKey(ctx context.Context, key string) (*entity.Task, error) {
	data, err := s.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, entity.ErrTaskNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения из Redis: %w", err)
	}

	var task entity.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("ошибка десериализации задачи: %w", err)
	}

	return &task, nil
}
