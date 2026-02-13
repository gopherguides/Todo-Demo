package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL         string
	Port                int
	ClerkSecretKey      string
	ClerkPublishableKey string
	SiteURL             string
}

func Load() (Config, error) {
	port := 8000
	if v := os.Getenv("PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("PORT must be a number: %w", err)
		}
		port = p
	}

	cfg := Config{
		DatabaseURL:         envOrDefault("DATABASE_URL", "data/todo.db"),
		Port:                port,
		ClerkSecretKey:      os.Getenv("CLERK_SECRET_KEY"),
		ClerkPublishableKey: os.Getenv("CLERK_PUBLISHABLE_KEY"),
		SiteURL:             envOrDefault("SITE_URL", "http://localhost:8000"),
	}

	if cfg.ClerkSecretKey == "" {
		return cfg, fmt.Errorf("CLERK_SECRET_KEY is required")
	}
	if cfg.ClerkPublishableKey == "" {
		return cfg, fmt.Errorf("CLERK_PUBLISHABLE_KEY is required")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
