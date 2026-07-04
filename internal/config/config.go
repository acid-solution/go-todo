package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port     string
	MySQLDSN string
}

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		Port:     getEnv("PORT", "8081"),
		MySQLDSN: getEnv("MYSQL_DSN", ""),
	}

	if cfg.MySQLDSN == "" {
		log.Fatal("MYSQL_DSN 未设置")
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}
