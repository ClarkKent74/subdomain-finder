package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"subdomain-finder/internal/entity"
	"subdomain-finder/internal/scanner"
	"subdomain-finder/internal/store"
	"subdomain-finder/pkg/validate"
)

type job struct {
	task  *entity.Task
	words []string // словарь для брутфорса, nil если используется встроенный
}

type FinderService struct {
	store       store.Store
	registry    scanner.Registry
	queue       chan job
	scanTimeout time.Duration
}

func NewFinderService(
	ctx context.Context,
	st store.Store,
	registry scanner.Registry,
	workerCount int,
	queueSize int,
	scanTimeout time.Duration,
) *FinderService {
	s := &FinderService{
		store:       st,
		registry:    registry,
		queue:       make(chan job, queueSize),
		scanTimeout: scanTimeout,
	}

	for i := 0; i < workerCount; i++ {
		go s.worker(ctx, i)
	}

	return s
}

func (s *FinderService) FindDomains(ctx context.Context, domain string, algorithm entity.Algorithm, words []string) error {
	if err := validate.Domain(domain); err != nil {
		return fmt.Errorf("%w: %s", entity.ErrInvalidDomain, err)
	}

	if _, ok := s.registry[algorithm]; !ok {
		return entity.ErrInvalidAlgorithm
	}

	if len(words) > 0 && algorithm != entity.AlgorithmBruteforce {
		return fmt.Errorf("словарь можно передавать только для алгоритма bruteforce")
	}

	now := time.Now()
	task := &entity.Task{
		Domain:    domain,
		Algorithm: algorithm,
		Status:    entity.StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.store.Create(ctx, task); err != nil {
		return err
	}

	select {
	case s.queue <- job{task: task, words: words}:
		return nil
	default:
		failTask(ctx, s.store, task, "очередь воркеров переполнена")
		return entity.ErrQueueFull
	}
}

func (s *FinderService) GetResult(ctx context.Context, domain string, algorithm entity.Algorithm) (*entity.Task, error) {
	if err := validate.Domain(domain); err != nil {
		return nil, fmt.Errorf("%w: %s", entity.ErrInvalidDomain, err)
	}

	if _, ok := s.registry[algorithm]; !ok {
		return nil, entity.ErrInvalidAlgorithm
	}

	return s.store.Get(ctx, domain, algorithm)
}

func (s *FinderService) worker(ctx context.Context, id int) {
	slog.Info("воркер запущен", "id", id)

	for {
		select {
		case j, ok := <-s.queue:
			if !ok {
				slog.Info("воркер завершён", "id", id)
				return
			}
			s.process(ctx, j)

		case <-ctx.Done():
			slog.Info("воркер остановлен по контексту", "id", id)
			return
		}
	}
}

func (s *FinderService) process(ctx context.Context, j job) {
	task := j.task

	slog.Info("начало сканирования",
		"domain", task.Domain,
		"algorithm", task.Algorithm,
	)

	task.Status = entity.StatusRunning
	task.UpdatedAt = time.Now()
	if err := s.store.Update(ctx, task); err != nil {
		slog.Error("не удалось обновить статус на Running",
			"domain", task.Domain,
			"error", err,
		)
		return
	}

	scanCtx, cancel := context.WithTimeout(ctx, s.scanTimeout)
	defer cancel()

	results, err := s.runScanner(scanCtx, task, j.words)

	if err != nil {
		if errors.Is(err, context.Canceled) {
			slog.Info("сканирование отменено", "domain", task.Domain)
			return
		}

		slog.Error("сканирование завершилось ошибкой",
			"domain", task.Domain,
			"algorithm", task.Algorithm,
			"error", err,
		)
		failTask(ctx, s.store, task, err.Error())
		return
	}

	task.Status = entity.StatusDone
	task.Results = results
	task.UpdatedAt = time.Now()
	if err := s.store.Update(ctx, task); err != nil {
		slog.Error("не удалось сохранить результаты",
			"domain", task.Domain,
			"error", err,
		)
		return
	}

	slog.Info("сканирование завершено",
		"domain", task.Domain,
		"algorithm", task.Algorithm,
		"найдено", len(results),
	)
}

func (s *FinderService) runScanner(ctx context.Context, task *entity.Task, words []string) ([]string, error) {
	sc := s.registry[task.Algorithm]

	if task.Algorithm == entity.AlgorithmBruteforce && len(words) > 0 {
		bf, ok := sc.(*scanner.BruteforceScanner)
		if !ok {
			return nil, fmt.Errorf("внутренняя ошибка: сканер bruteforce имеет неверный тип")
		}
		return bf.ScanWithWords(ctx, task.Domain, words)
	}

	return sc.Scan(ctx, task.Domain)
}

func failTask(ctx context.Context, st store.Store, task *entity.Task, reason string) {
	task.Status = entity.StatusFailed
	task.Error = reason
	task.UpdatedAt = time.Now()
	if err := st.Update(ctx, task); err != nil {
		slog.Error("не удалось обновить статус на Failed",
			"domain", task.Domain,
			"error", err,
		)
	}
}
