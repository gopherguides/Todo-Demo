package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL         string
	Port                string
	ClerkSecretKey      string
	ClerkPublishableKey string
	SiteURL             string
}

func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:         envOrDefault("DATABASE_URL", "data/todo.db"),
		Port:                envOrDefault("PORT", "8080"),
		ClerkSecretKey:      os.Getenv("CLERK_SECRET_KEY"),
		ClerkPublishableKey: os.Getenv("CLERK_PUBLISHABLE_KEY"),
		SiteURL:             envOrDefault("SITE_URL", "http://localhost:8080"),
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
