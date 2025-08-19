package config

import (
	"os"
	"strconv"
	"time"
)

// Config 應用程式設定結構
type Config struct {
	MongoDB  MongoDBConfig  `json:"mongodb"`
	API      APIConfig      `json:"api"`
	Logstash LogstashConfig `json:"logstash"`
}

// MongoDBConfig MongoDB設定
type MongoDBConfig struct {
	URI      string        `json:"uri"`
	Database string        `json:"database"`
	Timeout  time.Duration `json:"timeout"`
}

// APIConfig API設定
type APIConfig struct {
	SketchfabAPIKey string `json:"sketchfab_api_key"`
}

// LogstashConfig Logstash設定
type LogstashConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

// LoadConfig 載入設定
func LoadConfig() *Config {
	config := &Config{
		MongoDB: MongoDBConfig{
			URI:      getEnvOrDefault("MONGODB_URI", "mongodb://localhost:27017"),
			Database: getEnvOrDefault("MONGODB_DATABASE", "sketchfab_data"),
			Timeout:  getDurationEnvOrDefault("MONGODB_TIMEOUT", 10*time.Second),
		},
		API: APIConfig{
			SketchfabAPIKey: getEnvOrDefault("SKETCHFAB_API_KEY", ""),
		},
		Logstash: LogstashConfig{
			Host: getEnvOrDefault("LOGSTASH_HOST", "localhost"),
			Port: getEnvOrDefault("LOGSTASH_PORT", "5000"),
		},
	}

	return config
}

// getEnvOrDefault 取得環境變數或預設值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getDurationEnvOrDefault 取得時間間隔環境變數或預設值
func getDurationEnvOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if seconds, err := strconv.Atoi(value); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return defaultValue
}
