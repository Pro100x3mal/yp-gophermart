package configs

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

type ServerConfig struct {
	LogLevel    string
	RunAddr     string
	DatabaseURI string
	AccrualAddr string
	Secret      string
	TokenTTL    time.Duration
}

func GetConfig() (*ServerConfig, error) {
	var (
		cfg      ServerConfig
		tokenTTL int64
	)

	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.RunAddr, "a", "localhost:8080", "address of HTTP server")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "database PostgreSQL URI")
	flag.StringVar(&cfg.AccrualAddr, "r", "", "address of the accrual calculation system")
	flag.StringVar(&cfg.Secret, "s", "development-secret-change-me", "secret key for JWT")
	flag.Int64Var(&tokenTTL, "t", 24, "token TTL in hours")

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

	return &cfg, nil
}
