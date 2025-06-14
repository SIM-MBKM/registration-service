package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for our application
type Config struct {
	AppPort                   string
	DBHost                    string
	DBPort                    string
	DBUser                    string
	DBPassword                string
	DBName                    string
	RabbitMQHost              string
	RabbitMQPort              string
	RabbitMQUser              string
	RabbitMQPass              string
	NotificationQueue         string
	FrontendAllowedOrigins    []string
	FrontendAllowedReferers   []string
	FrontendRequireOrigin     bool
	FrontendBypassBrowsers    bool
	FrontendCustomHeader      string
	FrontendCustomHeaderValue string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		AppPort:                   getEnv("GOLANG_PORT", "8006"),
		DBHost:                    getEnv("DB_HOST", "localhost"),
		DBPort:                    getEnv("DB_PORT", "5432"),
		DBUser:                    getEnv("DB_USER", "postgres"),
		DBPassword:                getEnv("DB_PASSWORD", "postgres"),
		DBName:                    getEnv("DB_NAME", "notifications"),
		RabbitMQHost:              getEnv("RABBITMQ_HOST", "localhost"),
		RabbitMQPort:              getEnv("RABBITMQ_PORT", "5672"),
		RabbitMQUser:              getEnv("RABBITMQ_USER", "guest"),
		RabbitMQPass:              getEnv("RABBITMQ_PASS", "guest"),
		NotificationQueue:         getEnv("NOTIFICATION_QUEUE", "notification_queue"),
		FrontendAllowedOrigins:    getEnvAsSlice("FRONTEND_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		FrontendAllowedReferers:   getEnvAsSlice("FRONTEND_ALLOWED_REFERERS", []string{"localhost:3000"}),
		FrontendRequireOrigin:     getEnvAsBool("FRONTEND_REQUIRE_ORIGIN", true),
		FrontendBypassBrowsers:    getEnvAsBool("FRONTEND_BYPASS_BROWSERS", false),
		FrontendCustomHeader:      getEnv("FRONTEND_CUSTOM_HEADER", "X-Frontend-Request"),
		FrontendCustomHeaderValue: getEnv("FRONTEND_CUSTOM_HEADER_VALUE", "true"),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
