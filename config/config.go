package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                 string
	Host                 string
	Username             string
	Password             string
	HeartbeatIntervalMax uint16
	ChannelMax           uint16
	FrameMax             uint32
	Version              string
	Ssl                  bool
	QueueBufferSize      int
	WebServerPort        string
	JwtSecret            string
	LogLevel             string
}

// LoadConfig loads configuration from .env file, environment variables, or defaults
// Priority: environment variables > .env file > default values
func LoadConfig(version string) *Config {
	// Try to load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()

	return &Config{
		Port:                 getEnv("OTTERMQ_BROKER_PORT", "5672"),
		Host:                 getEnv("OTTERMQ_BROKER_HOST", ""),
		Username:             getEnv("OTTERMQ_USERNAME", "guest"),
		Password:             getEnv("OTTERMQ_PASSWORD", "guest"),
		HeartbeatIntervalMax: getEnvAsUint16("OTTERMQ_HEARTBEAT_INTERVAL", 60),
		ChannelMax:           getEnvAsUint16("OTTERMQ_CHANNEL_MAX", 2048),
		FrameMax:             getEnvAsUint32("OTTERMQ_FRAME_MAX", 131072),
		Ssl:                  getEnvAsBool("OTTERMQ_SSL", false),
		QueueBufferSize:      getEnvAsInt("OTTERMQ_QUEUE_BUFFER_SIZE", 100000),
		WebServerPort:        getEnv("OTTERMQ_WEB_PORT", "3000"),
		JwtSecret:            getEnv("OTTERMQ_JWT_SECRET", "secret"),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		Version:              version,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		fmt.Printf("Warning: Invalid value for %s: %s, using default: %d\n", key, valueStr, defaultValue)
		return defaultValue
	}
	return value
}

func getEnvAsUint16(key string, defaultValue uint16) uint16 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseUint(valueStr, 10, 16)
	if err != nil {
		fmt.Printf("Warning: Invalid value for %s: %s, using default: %d\n", key, valueStr, defaultValue)
		return defaultValue
	}
	return uint16(value)
}

func getEnvAsUint32(key string, defaultValue uint32) uint32 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseUint(valueStr, 10, 32)
	if err != nil {
		fmt.Printf("Warning: Invalid value for %s: %s, using default: %d\n", key, valueStr, defaultValue)
		return defaultValue
	}
	return uint32(value)
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		fmt.Printf("Warning: Invalid value for %s: %s, using default: %t\n", key, valueStr, defaultValue)
		return defaultValue
	}
	return value
}
