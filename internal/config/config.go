package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBURL        string
	TemporalHost string
	Port         string
}

func Load() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		DBURL:        getDBURL(),
		TemporalHost: getEnv("TEMPORAL_HOST", "localhost:7233"),
		Port:         getEnv("PORT", ":8080"),
	}

	if cfg.DBURL == "" {
		log.Fatal("error: Database configuration not found")
	}

	return cfg
}

func getDBURL() string {
	if dsn := os.Getenv("DB_URL"); dsn != "" {
		return dsn
	}

	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	name := getEnv("DB_NAME", "downloader")

	if user == "" || pass == "" {
		return ""
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, pass, host, port, name,
	)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
