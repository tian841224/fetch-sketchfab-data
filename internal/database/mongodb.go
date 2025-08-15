package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBConfig 包含MongoDB連線設定
type MongoDBConfig struct {
	URI      string
	Database string
	Timeout  time.Duration
}

// MongoDBClient 封裝MongoDB客戶端
type MongoDBClient struct {
	client   *mongo.Client
	database *mongo.Database
	config   *MongoDBConfig
}

// DefaultConfig 回傳預設的MongoDB設定
func DefaultConfig() *MongoDBConfig {
	return &MongoDBConfig{
		URI:      "mongodb://localhost:27017",
		Database: "sketchfab_data",
		Timeout:  10 * time.Second,
	}
}

// NewMongoDBClient 建立新的MongoDB客戶端
func NewMongoDBClient(config *MongoDBConfig) (*MongoDBClient, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 設定客戶端選項
	clientOptions := options.Client().ApplyURI(config.URI)

	// 建立連線
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("無法連線到MongoDB: %v", err)
	}

	// 測試連線
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("MongoDB連線測試失敗: %v", err)
	}

	log.Printf("成功連線到MongoDB: %s", config.URI)

	return &MongoDBClient{
		client:   client,
		database: client.Database(config.Database),
		config:   config,
	}, nil
}

// GetDatabase 取得資料庫實例
func (m *MongoDBClient) GetDatabase() *mongo.Database {
	return m.database
}

// GetCollection 取得指定的集合
func (m *MongoDBClient) GetCollection(collectionName string) *mongo.Collection {
	return m.database.Collection(collectionName)
}

// Close 關閉MongoDB連線
func (m *MongoDBClient) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), m.config.Timeout)
	defer cancel()

	err := m.client.Disconnect(ctx)
	if err != nil {
		return fmt.Errorf("關閉MongoDB連線時發生錯誤: %v", err)
	}

	log.Println("MongoDB連線已關閉")
	return nil
}

// IsConnected 檢查是否仍然連線
func (m *MongoDBClient) IsConnected() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.client.Ping(ctx, nil)
	return err == nil
}

// GetConnectionInfo 取得連線資訊
func (m *MongoDBClient) GetConnectionInfo() map[string]interface{} {
	return map[string]interface{}{
		"uri":       m.config.URI,
		"database":  m.config.Database,
		"timeout":   m.config.Timeout.String(),
		"connected": m.IsConnected(),
	}
}
