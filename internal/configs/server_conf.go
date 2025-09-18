package configs

import (
	"flag"
	"os"
)

type ServerConfig struct {
	LogLevel    string
	RunAddr     string
	DatabaseURI string
	AccrualAddr string
	Secret      string
}

func GetConfig() *ServerConfig {
	var cfg ServerConfig

	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.RunAddr, "a", "localhost:8080", "address of HTTP server")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "database PostgreSQL URI")
	flag.StringVar(&cfg.AccrualAddr, "r", "", "address of the accrual calculation system")
	flag.StringVar(&cfg.Secret, "s", "development-secret-change-me", "secret key for JWT")

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

	return &cfg
}
