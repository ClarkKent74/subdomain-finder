package configs

import "time"

type Config struct {
	Server  ServerConfig
	Workers WorkersConfig
	Store   StoreConfig
	Rate    RateConfig
	Sudomy  SudomyConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type WorkersConfig struct {
	PoolSize    int
	QueueSize   int
	ScanTimeout time.Duration
}

type StoreConfig struct {
	RedisURL string
	TaskTTL  time.Duration
	MaxTasks int
}

type RateConfig struct {
	RequestsPerMinute int
}

type SudomyConfig struct {
	Path              string
	ScanTimeout       time.Duration
	VirusTotalKey     string
	ShodanKey         string
	CensysKey         string
	SecurityTrailsKey string
}
