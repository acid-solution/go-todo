package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	MySQLDSN        string
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		Port:            getEnv("PORT", "8081"),
		MySQLDSN:        getEnv("MYSQL_DSN", ""),
		JWTSecret:       getEnv("JWT_SECRET", ""),
		AccessTokenTTL:  getDurationEnv("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL: getDurationEnv("REFRESH_TOKEN_TTL", 7*24*time.Hour),
	}

	if cfg.MySQLDSN == "" {
		log.Fatal("MYSQL_DSN 未设置")
	}

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET 未设置")
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

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		log.Fatalf("%s 格式错误: %v", key, err)
	}

	return duration
}
