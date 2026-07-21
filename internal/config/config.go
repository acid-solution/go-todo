package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	MySQLDSN         string
	JWTSecret        string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	TodoListCacheTTL time.Duration

	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		Port:             getEnv("PORT", "8081"),
		MySQLDSN:         getEnv("MYSQL_DSN", ""),
		JWTSecret:        getEnv("JWT_SECRET", ""),
		AccessTokenTTL:   getDurationEnv("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL:  getDurationEnv("REFRESH_TOKEN_TTL", 7*24*time.Hour),
		TodoListCacheTTL: getDurationEnv("TODO_LIST_CACHE_TTL", 5*time.Minute),

		RedisAddr:     getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getIntEnv("REDIS_DB", 0),
	}

	if cfg.MySQLDSN == "" {
		log.Fatal("MYSQL_DSN 未设置")
	}

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET 未设置")
	}

	if len(cfg.JWTSecret) < 32 {
		log.Fatal("JWT_SECRET 长度不能少于 32 字节")
	}

	if cfg.AccessTokenTTL <= 0 {
		log.Fatal("ACCESS_TOKEN_TTL 必须大于 0")
	}

	if cfg.RefreshTokenTTL <= 0 {
		log.Fatal("REFRESH_TOKEN_TTL 必须大于 0")
	}

	if cfg.TodoListCacheTTL <= 0 {
		log.Fatal("TODO_LIST_CACHE_TTL 必须大于 0")
	}

	if cfg.RefreshTokenTTL <= cfg.AccessTokenTTL {
		log.Fatal("REFRESH_TOKEN_TTL 必须大于 ACCESS_TOKEN_TTL")
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

func getIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	// strconv.Atoi 将字符串转换为整数，如果转换失败会返回错误
	number, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("%s 必须是整数: %v", key, err)
	}

	return number
}
