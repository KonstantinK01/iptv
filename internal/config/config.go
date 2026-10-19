package config

import (
	"os"
)

type Config struct {
	ListenAddr    string
	DatabaseURL   string
	StreamBaseURL string
}

func New() Config {
	return Config{
		ListenAddr:    env("LISTEN_ADDR", ":8080"),
		DatabaseURL:   env("DATABASE_URL", "postgres://iptv:iptv@localhost:5432/iptv?sslmode=disable"),
		StreamBaseURL: env("STREAM_BASE_URL", "http://stream.example.com"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
