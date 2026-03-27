package configs

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func Load() *Config {
	_ = godotenv.Load(".env")

	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  getDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Workers: WorkersConfig{
			PoolSize:    getInt("WORKER_POOL_SIZE", 10),
			QueueSize:   getInt("WORKER_QUEUE_SIZE", 100),
			ScanTimeout: getDuration("WORKER_SCAN_TIMEOUT", 5*time.Minute),
		},
		Store: StoreConfig{
			RedisURL: getEnv("REDIS_URL", "redis://localhost:6379"),
			TaskTTL:  getDuration("STORE_TASK_TTL", 24*time.Hour),
			MaxTasks: getInt("STORE_MAX_TASKS", 10000),
		},
		Rate: RateConfig{
			RequestsPerMinute: getInt("RATE_LIMIT_RPM", 20),
		},
		Sudomy: SudomyConfig{
			Path:              getEnv("SUDOMY_PATH", "/usr/lib/sudomy/sudomy"),
			ScanTimeout:       getDuration("SUDOMY_SCAN_TIMEOUT", 30*time.Minute),
			VirusTotalKey:     getEnv("SUDOMY_VIRUSTOTAL_KEY", ""),
			ShodanKey:         getEnv("SUDOMY_SHODAN_KEY", ""),
			CensysKey:         getEnv("SUDOMY_CENSYS_KEY", ""),
			SecurityTrailsKey: getEnv("SUDOMY_SECURITYTRAILS_KEY", ""),
		},
	}
}

func (c *Config) Validate() error {
	if c.Workers.PoolSize < 1 || c.Workers.PoolSize > 100 {
		return fmt.Errorf("WORKER_POOL_SIZE должен быть от 1 до 100, получено: %d", c.Workers.PoolSize)
	}
	if c.Workers.QueueSize < 1 || c.Workers.QueueSize > 10000 {
		return fmt.Errorf("WORKER_QUEUE_SIZE должен быть от 1 до 10000, получено: %d", c.Workers.QueueSize)
	}
	if c.Store.MaxTasks < 1 || c.Store.MaxTasks > 1000000 {
		return fmt.Errorf("STORE_MAX_TASKS должен быть от 1 до 1000000, получено: %d", c.Store.MaxTasks)
	}
	if c.Rate.RequestsPerMinute < 1 || c.Rate.RequestsPerMinute > 1000 {
		return fmt.Errorf("RATE_LIMIT_RPM должен быть от 1 до 1000, получено: %d", c.Rate.RequestsPerMinute)
	}
	return nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return n
}

func getDuration(key string, defaultVal time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return defaultVal
	}
	return d
}
