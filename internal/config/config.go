package config

import (
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var (
	ErrMissingSecret   = errors.New("APP_SECRET is required")
	ErrSecretTooShort  = errors.New("APP_SECRET must be at least 32 bytes")
	ErrInvalidDatabase = errors.New("DATABASE_URL is invalid")
	ErrInvalidRedis    = errors.New("REDIS_URL is invalid")
)

type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	R2        R2Config
	Email     EmailConfig
	SMS       SMSConfig
	OAuth     OAuthConfig
	Telemetry TelemetryConfig
	Features  FeaturesConfig
}

type AppConfig struct {
	Env         string
	Port        string
	Secret      []byte
	BaseURL     string
	FrontendURL string
	Host        string
}

type DatabaseConfig struct {
	URL      string
	MaxConns int32
	MaxIdle  int32
	MinConns int32
	ConnTTL  time.Duration
}

type RedisConfig struct {
	URL         string
	Password    string
	DB          int
	PoolSize    int
	PoolTimeout time.Duration
}

type R2Config struct {
	AccountID string
	AccessKey string
	SecretKey string
	Bucket    string
	PublicURL string
}

type EmailConfig struct {
	Host string
	Port int
	User string
	Pass string
	From string
}

type SMSConfig struct {
	AccountSID string
	AuthToken  string
	FromNumber string
}

type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	AppleClientID      string
	AppleClientSecret  string
	AppleTeamID        string
	AppleKeyID         string
	ApplePrivateKey    string
}

type TelemetryConfig struct {
	OTLPEndpoint   string
	PrometheusPort string
}

type FeaturesConfig struct {
	EmailSync bool
	SMS       bool
	WhatsApp  bool
	APIKeys   bool
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error loading .env: %w", err)
	}

	cfg := &Config{
		App: AppConfig{
			Env:         getEnv("APP_ENV", "development"),
			Port:        getEnv("APP_PORT", "8080"),
			Secret:      decodeHexSecret(getEnv("APP_SECRET", "")),
			BaseURL:     getEnv("APP_BASE_URL", "http://localhost:8080"),
			FrontendURL: getEnv("APP_FRONTEND_URL", "http://localhost:4321"),
			Host:        getEnv("APP_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			URL:      getEnv("DATABASE_URL", ""),
			MaxConns: int32(getEnvInt("DATABASE_MAX_CONNS", 25)),
			MaxIdle:  int32(getEnvInt("DATABASE_MAX_IDLE", 5)),
			MinConns: int32(getEnvInt("DATABASE_MIN_CONNS", 5)),
			ConnTTL:  getEnvDuration("DATABASE_CONN_TTL", 30*time.Minute),
		},
		Redis: RedisConfig{
			URL:         getEnv("REDIS_URL", ""),
			PoolSize:    getEnvInt("REDIS_POOL_SIZE", 10),
			PoolTimeout: getEnvDuration("REDIS_POOL_TIMEOUT", 5*time.Second),
			DB:          getEnvInt("REDIS_DB", 0),
		},
		R2: R2Config{
			AccountID: getEnv("R2_ACCOUNT_ID", ""),
			AccessKey: getEnv("R2_ACCESS_KEY", ""),
			SecretKey: getEnv("R2_SECRET_KEY", ""),
			Bucket:    getEnv("R2_BUCKET", "oscar"),
			PublicURL: getEnv("R2_PUBLIC_URL", ""),
		},
		Email: EmailConfig{
			Host: getEnv("SMTP_HOST", "localhost"),
			Port: getEnvInt("SMTP_PORT", 587),
			User: getEnv("SMTP_USER", ""),
			Pass: getEnv("SMTP_PASS", ""),
			From: getEnv("SMTP_FROM", "noreply@oscar.local"),
		},
		SMS: SMSConfig{
			AccountSID: getEnv("TWILIO_ACCOUNT_SID", ""),
			AuthToken:  getEnv("TWILIO_AUTH_TOKEN", ""),
			FromNumber: getEnv("TWILIO_FROM_NUMBER", ""),
		},
		Telemetry: TelemetryConfig{
			OTLPEndpoint:   getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
			PrometheusPort: getEnv("PROMETHEUS_PORT", "9090"),
		},
		Features: FeaturesConfig{
			EmailSync: getEnvBool("FEATURE_EMAIL_SYNC", false),
			SMS:       getEnvBool("FEATURE_SMS", false),
			WhatsApp:  getEnvBool("FEATURE_WHATSAPP", false),
			APIKeys:   getEnvBool("FEATURE_API_KEYS", true),
		},
		OAuth: OAuthConfig{
			GoogleClientID:     getEnv("OAUTH_GOOGLE_CLIENT_ID", ""),
			GoogleClientSecret: getEnv("OAUTH_GOOGLE_CLIENT_SECRET", ""),
			AppleClientID:      getEnv("OAUTH_APPLE_CLIENT_ID", ""),
			AppleClientSecret:  getEnv("OAUTH_APPLE_CLIENT_SECRET", ""),
			AppleTeamID:        getEnv("OAUTH_APPLE_TEAM_ID", ""),
			AppleKeyID:         getEnv("OAUTH_APPLE_KEY_ID", ""),
			ApplePrivateKey:    getEnv("OAUTH_APPLE_PRIVATE_KEY", ""),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if len(c.App.Secret) == 0 {
		return ErrMissingSecret
	}
	if len(c.App.Secret) < 32 {
		return ErrSecretTooShort
	}
	if c.App.Env == "production" && subtle.ConstantTimeCompare(c.App.Secret, []byte("change-me-in-production-32bytes!!")) == 1 {
		return errors.New("APP_SECRET must be changed in production")
	}

	if _, err := url.Parse(c.Database.URL); err != nil || c.Database.URL == "" {
		return fmt.Errorf("%w: %s", ErrInvalidDatabase, c.Database.URL)
	}

	if _, err := url.Parse(c.Redis.URL); err != nil || c.Redis.URL == "" {
		return fmt.Errorf("%w: %s", ErrInvalidRedis, c.Redis.URL)
	}

	return nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		return strings.ToLower(val) == "true" || val == "1"
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			return duration
		}
	}
	return defaultVal
}

func decodeHexSecret(hexStr string) []byte {
	if hexStr == "" {
		return nil
	}
	decoded, err := hex.DecodeString(hexStr)
	if err != nil {
		return []byte(hexStr)
	}
	return decoded
}
