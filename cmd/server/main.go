// @title           Subdomain Finder API
// @version         1.0
// @description     Сервис поиска поддоменов.
// @host            localhost:8080
// @BasePath        /

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "subdomain-finder/docs"
	"subdomain-finder/internal/api"
	"subdomain-finder/internal/configs"
	"subdomain-finder/internal/entity"
	"subdomain-finder/internal/scanner"
	"subdomain-finder/internal/service"
	"subdomain-finder/internal/store"
)

func main() {
	cfg := configs.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("некорректная конфигурация: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	activeTTL := cfg.Workers.ScanTimeout * 2
	st, err := store.NewRedisStore(ctx,
		cfg.Store.RedisURL,
		cfg.Store.TaskTTL,
		activeTTL,
		cfg.Store.MaxTasks,
	)
	if err != nil {
		log.Fatalf("не удалось подключиться к Redis: %v", err)
	}
	defer st.Close()

	registry := scanner.Registry{
		entity.AlgorithmPassive: scanner.NewPassiveScanner(
			cfg.Scanner.PassiveTimeout,
			cfg.Scanner.CircuitBreakerThreshold,
			cfg.Scanner.CircuitBreakerTimeout,
		),
		entity.AlgorithmBruteforce: scanner.NewBruteforceScanner(
			cfg.DNS.Resolver,
			cfg.DNS.WorkerCount,
			cfg.DNS.QueryTimeout,
		),
		entity.AlgorithmZoneTransfer: scanner.NewZoneTransferScanner(
			cfg.DNS.QueryTimeout,
		),
	}

	svc := service.NewFinderService(
		ctx,
		st,
		registry,
		cfg.Workers.PoolSize,
		cfg.Workers.QueueSize,
		cfg.Workers.ScanTimeout,
	)

	handler := api.NewHandler(svc)
	router := api.NewRouter(handler, cfg.Rate.RequestsPerMinute)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		slog.Info("сервер запущен", "адрес", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("ошибка сервера: %v", err)
		}
	}()

	fmt.Printf("Сервис запущен на http://localhost:%s\n", cfg.Server.Port)
	fmt.Printf("Swagger UI:  http://localhost:%s/swagger/index.html\n", cfg.Server.Port)
	fmt.Println()
	fmt.Println("Для остановки - Ctrl+C")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("получен сигнал остановки, завершаем работу...")

	// Даём 30 секунд на завершение активных запросов
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Останавливаем HTTP сервер — новые запросы не принимаем
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("ошибка graceful shutdown", "error", err)
	}

	// Отменяем контекст — останавливаем воркеры
	cancel()

	slog.Info("сервис остановлен")
}
