package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// this holds the configuration for the application.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Paste    PasteConfig    `mapstructure:"paste"`
	Security SecurityConfig `mapstructure:"security"`
	S3       S3Config       `mapstructure:"s3"`
}

// holds server-related settings.
type ServerConfig struct {
	Port    int    `mapstructure:"port"`
	BaseURL string `mapstructure:"base_url"`
}

// holds database-related settings.
type DatabaseConfig struct {
	ConnectionString string `mapstructure:"connection_string"`
}

// holds paste-related settings.
type PasteConfig struct {
	IDLength                 int `mapstructure:"id_length"`
	DefaultExpirationMinutes int `mapstructure:"default_expiration_minutes"`
	MaxContentSizeBytes      int `mapstructure:"max_content_size_bytes"`
}

// holds security-related settings.
type SecurityConfig struct {
	BcryptCost         int `mapstructure:"bcrypt_cost"`
	RateLimitPerSecond int `mapstructure:"rate_limit_per_second"`
}

type S3Config struct {
	Bucket          string `mapstructure:"bucket"`
	Region          string `mapstructure:"region"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
}

// this will read configuration from file or environment variables.
func LoadConfig() (*Config, error) {
	viper.AddConfigPath("./configs") //  config.toml
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	viper.SetDefault("server.port", 8080)
	viper.SetDefault("paste.id_length", 8)
	viper.SetDefault("paste.default_expiration_minutes", 0)
	viper.SetDefault("paste.max_content_size_bytes", 1048576) // 1MB
	viper.SetDefault("security.bcrypt_cost", 12)
	viper.SetDefault("security.rate_limit_per_second", 5)

	// env variables can override config file settings
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // allows env vars like SERVER_PORT

	viper.BindEnv("database.connection_string", "DATABASE_CONNECTION_STRING")
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.base_url", "BASE_URL")

	// Bind S3 configuration to AWS environment variables
	viper.BindEnv("s3.bucket", "AWS_S3_BUCKET")
	viper.BindEnv("s3.region", "AWS_REGION")
	viper.BindEnv("s3.access_key_id", "AWS_ACCESS_KEY_ID")
	viper.BindEnv("s3.secret_access_key", "AWS_SECRET_ACCESS_KEY")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// config file not found; using defaults and env vars
			fmt.Println("Config file not found, using defaults and environment variables.")
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// validation for critical config
	if cfg.Database.ConnectionString == "" {
		return nil, fmt.Errorf("database connection string cannot be empty")
	}
	if cfg.Paste.IDLength <= 0 {
		return nil, fmt.Errorf("paste ID length must be positive")
	}

	if cfg.S3.Bucket == "" {
		fmt.Println("WARNING: S3 bucket not configured - file uploads will fail")
	}
	if cfg.S3.AccessKeyID == "" || cfg.S3.SecretAccessKey == "" {
		fmt.Println("WARNING: S3 credentials not configured - file uploads will fail")
	}
	if cfg.S3.Region == "" {
		fmt.Println("WARNING: S3 region not configured - file uploads will fail")
	}

	return &cfg, nil
}

// GetExpirationTime calculates the expiration time based on minutes.
// Returns nil if minutes is 0 (never expires).
func (p PasteConfig) GetExpirationTime(minutes int) *time.Time {
	if minutes <= 0 {
		return nil
	}
	expiresAt := time.Now().Add(time.Duration(minutes) * time.Minute)
	return &expiresAt
}
