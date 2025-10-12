package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server        ServerConfig
	Database      DatabaseConfig
	Email         EmailConfig
	Auth          AuthConfig
	CORS          CORSConfig
	ExternalAPIs  ExternalAPIsConfig
	Redis         RedisConfig
	Elasticsearch ElasticsearchConfig
	Upload        UploadConfig
	RateLimit     RateLimitConfig
	Logging       LoggingConfig
	Seeding       SeedingConfig
	BrokerSync    BrokerSyncConfig
	Encryption    EncryptionConfig
}

type ServerConfig struct {
	Port        string
	Host        string
	FrontendURL string // URL of the frontend application for email links
}

type DatabaseConfig struct {
	URL  string
	Host string
	Port int
	User string
	Pass string
	Name string
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

type AuthConfig struct {
	JWTSecret      string
	JWTExpiration  string
	CookieDomain   string
	CookieSecure   bool
	CookieSameSite string // "strict", "lax", "none"
	CookieHTTPOnly bool
	CookieMaxAge   int // in seconds
}

type CORSConfig struct {
	Origins []string
}

type ExternalAPIsConfig struct {
	GoogleClientID     string
	GoogleClientSecret string

	// Financial Data APIs
	AlphaVantageAPIKey          string
	YahooFinanceAPIKey          string
	FinancialModelingPrepAPIKey string

	// Exchange Rate APIs
	ExchangeRateAPIKey  string
	CurrencyLayerAPIKey string

	// Crypto APIs
	CoinGeckoAPIKey     string
	CoinMarketCapAPIKey string
}

type RedisConfig struct {
	URL string
}

type ElasticsearchConfig struct {
	URL string
}

type UploadConfig struct {
	MaxSize      int64
	AllowedTypes []string
}

type RateLimitConfig struct {
	Requests int
	Window   string
}

type LoggingConfig struct {
	Level  string
	Format string
}

type SeedingConfig struct {
	AdminEmail    string
	AdminPassword string
}

type BrokerSyncConfig struct {
	Enabled       bool
	IntervalMin   int // Interval in minutes
	MaxConcurrent int
	TimeoutMin    int // Timeout per sync in minutes
}

type EncryptionConfig struct {
	Key string // Must be 32 bytes for AES-256
}

// Load initializes and loads configuration using Viper
func Load() *Config {
	// Initialize Viper
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./server")
	viper.AddConfigPath("../")

	// Enable automatic environment variable reading
	viper.AutomaticEnv()

	// Set environment variable prefix (optional)
	viper.SetEnvPrefix("")

	// Replace dots and dashes with underscores in env keys
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Set default values
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("Warning: .env file not found, using environment variables and defaults")
		} else {
			log.Printf("Error reading config file: %v", err)
		}
	} else {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}

	// Build config from Viper
	config := &Config{
		Server: ServerConfig{
			Port:        viper.GetString("PORT"),
			Host:        viper.GetString("HOST"),
			FrontendURL: viper.GetString("FRONTEND_URL"),
		},
		Database: DatabaseConfig{
			URL:  viper.GetString("DATABASE_URL"),
			Host: viper.GetString("DB_HOST"),
			Port: viper.GetInt("DB_PORT"),
			User: viper.GetString("DB_USER"),
			Pass: viper.GetString("DB_PASSWORD"),
			Name: viper.GetString("DB_NAME"),
		},
		Email: EmailConfig{
			SMTPHost:     viper.GetString("SMTP_HOST"),
			SMTPPort:     viper.GetInt("SMTP_PORT"),
			SMTPUsername: viper.GetString("SMTP_USERNAME"),
			SMTPPassword: viper.GetString("SMTP_PASSWORD"),
			FromEmail:    viper.GetString("FROM_EMAIL"),
			FromName:     viper.GetString("FROM_NAME"),
		},
		Auth: AuthConfig{
			JWTSecret:      viper.GetString("JWT_SECRET"),
			JWTExpiration:  viper.GetString("JWT_EXPIRATION"),
			CookieDomain:   viper.GetString("COOKIE_DOMAIN"),
			CookieSecure:   viper.GetBool("COOKIE_SECURE"),
			CookieSameSite: viper.GetString("COOKIE_SAME_SITE"),
			CookieHTTPOnly: viper.GetBool("COOKIE_HTTP_ONLY"),
			CookieMaxAge:   viper.GetInt("COOKIE_MAX_AGE"),
		},
		CORS: CORSConfig{
			Origins: viper.GetStringSlice("CORS_ORIGINS"),
		},
		ExternalAPIs: ExternalAPIsConfig{
			GoogleClientID:     viper.GetString("GOOGLE_CLIENT_ID"),
			GoogleClientSecret: viper.GetString("GOOGLE_CLIENT_SECRET"),

			// Financial Data APIs
			AlphaVantageAPIKey:          viper.GetString("ALPHA_VANTAGE_API_KEY"),
			YahooFinanceAPIKey:          viper.GetString("YAHOO_FINANCE_API_KEY"),
			FinancialModelingPrepAPIKey: viper.GetString("FINANCIAL_MODELING_PREP_API_KEY"),

			// Exchange Rate APIs
			ExchangeRateAPIKey:  viper.GetString("EXCHANGE_RATE_API_KEY"),
			CurrencyLayerAPIKey: viper.GetString("CURRENCY_LAYER_API_KEY"),

			// Crypto APIs
			CoinGeckoAPIKey:     viper.GetString("COINGECKO_API_KEY"),
			CoinMarketCapAPIKey: viper.GetString("COINMARKETCAP_API_KEY"),
		},
		Redis: RedisConfig{
			URL: viper.GetString("REDIS_URL"),
		},
		Elasticsearch: ElasticsearchConfig{
			URL: viper.GetString("ELASTICSEARCH_URL"),
		},
		Upload: UploadConfig{
			MaxSize:      viper.GetInt64("UPLOAD_MAX_SIZE"),
			AllowedTypes: viper.GetStringSlice("UPLOAD_ALLOWED_TYPES"),
		},
		RateLimit: RateLimitConfig{
			Requests: viper.GetInt("RATE_LIMIT_REQUESTS"),
			Window:   viper.GetString("RATE_LIMIT_WINDOW"),
		},
		Logging: LoggingConfig{
			Level:  viper.GetString("LOG_LEVEL"),
			Format: viper.GetString("LOG_FORMAT"),
		},
		Seeding: SeedingConfig{
			AdminEmail:    viper.GetString("ADMIN_EMAIL"),
			AdminPassword: viper.GetString("ADMIN_PASSWORD"),
		},
		BrokerSync: BrokerSyncConfig{
			Enabled:       viper.GetBool("BROKER_SYNC_ENABLED"),
			IntervalMin:   viper.GetInt("BROKER_SYNC_INTERVAL_MIN"),
			MaxConcurrent: viper.GetInt("BROKER_SYNC_MAX_CONCURRENT"),
			TimeoutMin:    viper.GetInt("BROKER_SYNC_TIMEOUT_MIN"),
		},
		Encryption: EncryptionConfig{
			Key: viper.GetString("ENCRYPTION_KEY"),
		},
	}

	return config
}

// setDefaults sets default values for all configuration options
func setDefaults() {
	// Server Configuration
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("HOST", "localhost")
	viper.SetDefault("FRONTEND_URL", "http://localhost:3000")
	viper.SetDefault("GIN_MODE", "debug")

	// Database Configuration
	viper.SetDefault("DATABASE_URL", "")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_USER", "finance_user")
	viper.SetDefault("DB_PASSWORD", "finance_password")
	viper.SetDefault("DB_NAME", "finance_dss")

	// JWT Configuration
	viper.SetDefault("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production")
	viper.SetDefault("JWT_EXPIRATION", "24h")

	// Cookie Configuration
	viper.SetDefault("COOKIE_DOMAIN", "")          // Empty = current domain
	viper.SetDefault("COOKIE_SECURE", false)       // Set true in production with HTTPS
	viper.SetDefault("COOKIE_SAME_SITE", "lax")    // "strict", "lax", or "none"
	viper.SetDefault("COOKIE_HTTP_ONLY", true)     // Always true for security
	viper.SetDefault("COOKIE_MAX_AGE", 7*24*60*60) // 7 days in seconds

	// Email Configuration
	viper.SetDefault("SMTP_HOST", "smtp.gmail.com")
	viper.SetDefault("SMTP_PORT", 587)
	viper.SetDefault("SMTP_USERNAME", "")
	viper.SetDefault("SMTP_PASSWORD", "")
	viper.SetDefault("FROM_EMAIL", "noreply@personalfinancedss.com")
	viper.SetDefault("FROM_NAME", "Personal Finance DSS")

	// CORS Configuration
	viper.SetDefault("CORS_ORIGINS", []string{"http://localhost:3000", "http://127.0.0.1:3000"})

	// External APIs
	viper.SetDefault("GOOGLE_CLIENT_ID", "")
	viper.SetDefault("GOOGLE_CLIENT_SECRET", "")

	// Financial Data APIs
	viper.SetDefault("ALPHA_VANTAGE_API_KEY", "")
	viper.SetDefault("YAHOO_FINANCE_API_KEY", "")
	viper.SetDefault("FINANCIAL_MODELING_PREP_API_KEY", "")

	// Exchange Rate APIs
	viper.SetDefault("EXCHANGE_RATE_API_KEY", "")
	viper.SetDefault("CURRENCY_LAYER_API_KEY", "")

	// Crypto APIs
	viper.SetDefault("COINGECKO_API_KEY", "")
	viper.SetDefault("COINMARKETCAP_API_KEY", "")

	// Redis Configuration
	viper.SetDefault("REDIS_URL", "redis://localhost:6379")

	// Elasticsearch Configuration
	viper.SetDefault("ELASTICSEARCH_URL", "http://localhost:9200")

	// File Upload Configuration
	viper.SetDefault("UPLOAD_MAX_SIZE", 10485760) // 10MB
	viper.SetDefault("UPLOAD_ALLOWED_TYPES", []string{"image/jpeg", "image/png", "application/pdf"})

	// Rate Limiting
	viper.SetDefault("RATE_LIMIT_REQUESTS", 100)
	viper.SetDefault("RATE_LIMIT_WINDOW", "1m")

	// Logging
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FORMAT", "json")

	// Database Seeding
	viper.SetDefault("ADMIN_EMAIL", "admin@example.com")
	viper.SetDefault("ADMIN_PASSWORD", "Admin@123")

	// Broker Sync Configuration
	viper.SetDefault("BROKER_SYNC_ENABLED", true)
	viper.SetDefault("BROKER_SYNC_INTERVAL_MIN", 1)
	viper.SetDefault("BROKER_SYNC_MAX_CONCURRENT", 5)
	viper.SetDefault("BROKER_SYNC_TIMEOUT_MIN", 2)

	// Encryption Configuration
	// IMPORTANT: Change this in production! Must be exactly 32 bytes for AES-256
	viper.SetDefault("ENCRYPTION_KEY", "dev-key-32bytes-change-in-prod!!")
}
