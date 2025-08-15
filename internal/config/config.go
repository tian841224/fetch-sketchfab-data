package config

import (
	"os"
	"strconv"
	"time"
)

// Config 應用程式設定結構
type Config struct {
	MongoDB MongoDBConfig `json:"mongodb"`
	API     APIConfig     `json:"api"`
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
