package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server       ServerConfig
	Database     DatabaseConfig
	Logging      LoggingConfig
	ServerLimits ServerLimitsConfig
	App          AppConfig
}

type ServerConfig struct {
	Port       string
	Host       string
	CORSOrigins string // Comma-separated list of allowed CORS origins
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	Charset  string
}

type LoggingConfig struct {
	Directory         string
	RetentionDays     int
	CompressAfterDays int
	Level             string
}

type ServerLimitsConfig struct {
	RequestTimeoutSeconds int     // Request timeout in seconds
	RateLimitRPS          float64 // Rate limit requests per second
	RateLimitBurst        int     // Rate limit burst size
	MaxRequestSizeMB      int     // Max request size in MB
}

type AppConfig struct {
	IsProduction bool
}

func Load() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	retentionDays := 30
	if days := getEnv("LOG_RETENTION_DAYS", "30"); days != "" {
		if parsed, err := strconv.Atoi(days); err == nil {
			retentionDays = parsed
		}
	}

	compressAfterDays := 7
	if days := getEnv("LOG_COMPRESS_AFTER_DAYS", "7"); days != "" {
		if parsed, err := strconv.Atoi(days); err == nil {
			compressAfterDays = parsed
		}
	}

	// Server limits config
	requestTimeout := 30
	if timeout := getEnv("REQUEST_TIMEOUT_SECONDS", "30"); timeout != "" {
		if parsed, err := strconv.Atoi(timeout); err == nil {
			requestTimeout = parsed
		}
	}

	rateLimitRPS := 100.0
	if rps := getEnv("RATE_LIMIT_RPS", "100"); rps != "" {
		if parsed, err := strconv.ParseFloat(rps, 64); err == nil {
			rateLimitRPS = parsed
		}
	}

	rateLimitBurst := 200
	if burst := getEnv("RATE_LIMIT_BURST", "200"); burst != "" {
		if parsed, err := strconv.Atoi(burst); err == nil {
			rateLimitBurst = parsed
		}
	}

	maxRequestSizeMB := 10
	if size := getEnv("MAX_REQUEST_SIZE_MB", "10"); size != "" {
		if parsed, err := strconv.Atoi(size); err == nil {
			maxRequestSizeMB = parsed
		}
	}

	// Determine if production mode
	env := getEnv("ENV", "development")
	isProduction := env == "production" || env == "prod"

	cfg := &Config{
		Server: ServerConfig{
			Port:        getEnv("SERVER_PORT", "8085"),
			Host:        getEnv("SERVER_HOST", "0.0.0.0"),
			CORSOrigins: getEnv("CORS_ORIGINS", ""),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "clean_architecture"),
			Charset:  getEnv("DB_CHARSET", "utf8mb4"),
		},
		Logging: LoggingConfig{
			Directory:         getEnv("LOG_DIRECTORY", "./logs"),
			RetentionDays:     retentionDays,
			CompressAfterDays: compressAfterDays,
			Level:             getEnv("LOG_LEVEL", "info"),
		},
		ServerLimits: ServerLimitsConfig{
			RequestTimeoutSeconds: requestTimeout,
			RateLimitRPS:          rateLimitRPS,
			RateLimitBurst:        rateLimitBurst,
			MaxRequestSizeMB:      maxRequestSizeMB,
		},
		App: AppConfig{
			IsProduction: isProduction,
		},
	}

	return cfg, nil
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.Charset)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
