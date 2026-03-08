package configs

import "time"

// Config — полная конфигурация приложения.
type Config struct {
	Server  ServerConfig
	Workers WorkersConfig
	Store   StoreConfig
	Rate    RateConfig
	DNS     DNSConfig
	Scanner ScannerConfig
}

// ServerConfig — параметры HTTP сервера.
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// WorkersConfig — параметры пула воркеров.
type WorkersConfig struct {
	PoolSize    int
	QueueSize   int
	ScanTimeout time.Duration
}

// StoreConfig — параметры хранилища задач.
type StoreConfig struct {
	RedisURL string
	TaskTTL  time.Duration
	MaxTasks int
}

// RateConfig — параметры rate limiter.
type RateConfig struct {
	RequestsPerMinute int
}

// DNSConfig — параметры DNS запросов для брутфорса.
type DNSConfig struct {
	Resolver     string
	QueryTimeout time.Duration
	WorkerCount  int
}

// ScannerConfig — параметры внешних сканеров.
type ScannerConfig struct {
	PassiveTimeout          time.Duration
	CircuitBreakerThreshold int
	CircuitBreakerTimeout   time.Duration
}
