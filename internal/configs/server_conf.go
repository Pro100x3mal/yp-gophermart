package configs

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

type ServerConfig struct {
	LogLevel     string
	RunAddr      string
	DatabaseURI  string
	AccrualAddr  string
	Secret       string
	BatchSize    int
	RateLimit    int
	TokenTTL     time.Duration
	PollInterval time.Duration
}

func GetConfig() (*ServerConfig, error) {
	var (
		cfg          ServerConfig
		tokenTTL     int64
		pollInterval int64
	)

	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.RunAddr, "a", "localhost:8080", "address of HTTP server")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "database PostgreSQL URI")
	flag.StringVar(&cfg.AccrualAddr, "r", "", "address of the accrual calculation system")
	flag.StringVar(&cfg.Secret, "s", "development-secret-change-me", "secret key for JWT")
	flag.IntVar(&cfg.BatchSize, "b", 10, "batch size for accrual requests")
	flag.IntVar(&cfg.RateLimit, "n", 5, "rate limit for accrual requests")
	flag.Int64Var(&tokenTTL, "t", 24, "token TTL in hours")
	flag.Int64Var(&pollInterval, "i", 1, "poll interval in seconds")

	flag.Parse()

	if envLogLevel, ok := os.LookupEnv("LOG_LEVEL"); ok && envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}

	if envRunAddr, ok := os.LookupEnv("RUN_ADDRESS"); ok && envRunAddr != "" {
		cfg.RunAddr = envRunAddr
	}

	if envDatabaseDSN, ok := os.LookupEnv("DATABASE_URI"); ok && envDatabaseDSN != "" {
		cfg.DatabaseURI = envDatabaseDSN
	}

	if envAccrualAddr, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok && envAccrualAddr != "" {
		cfg.AccrualAddr = envAccrualAddr
	}

	if envSecret, ok := os.LookupEnv("SECRET"); ok && envSecret != "" {
		cfg.Secret = envSecret
	}

	if envBatchSize, ok := os.LookupEnv("BATCH_SIZE"); ok && envBatchSize != "" {
		var err error
		cfg.BatchSize, err = strconv.Atoi(envBatchSize)
		if err != nil {
			return nil, fmt.Errorf("failed to parse BATCH_SIZE value %q to integer: %w", envBatchSize, err)
		}
		if cfg.BatchSize <= 0 {
			return nil, fmt.Errorf("invalid BATCH_SIZE value %q: must be positive", envBatchSize)
		}
	}

	if envRateLimit, ok := os.LookupEnv("RATE_LIMIT"); ok && envRateLimit != "" {
		var err error
		cfg.RateLimit, err = strconv.Atoi(envRateLimit)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RATE_LIMIT value %q to integer: %w", envRateLimit, err)
		}
		if cfg.RateLimit <= 0 {
			return nil, fmt.Errorf("invalid RATE_LIMIT value %q: must be positive", envRateLimit)
		}
	}

	if envTokenTTL, ok := os.LookupEnv("TOKEN_TTL"); ok && envTokenTTL != "" {
		var err error
		tokenTTL, err = strconv.ParseInt(envTokenTTL, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse TOKEN_TTL value %q to integer: %w", envTokenTTL, err)
		}
		if tokenTTL <= 0 {
			return nil, fmt.Errorf("invalid TOKEN_TTL value %q: must be positive", envTokenTTL)
		}
	}
	cfg.TokenTTL = time.Duration(tokenTTL) * time.Hour

	if envPollInterval, ok := os.LookupEnv("POLL_INTERVAL"); ok && envPollInterval != "" {
		var err error
		pollInterval, err = strconv.ParseInt(envPollInterval, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse POLL_INTERVAL value %q to integer: %w", envPollInterval, err)
		}
		if pollInterval <= 0 {
			return nil, fmt.Errorf("invalid POLL_INTERVAL value %q: must be positive", envPollInterval)
		}
	}
	cfg.PollInterval = time.Duration(pollInterval) * time.Second

	return &cfg, nil
}
