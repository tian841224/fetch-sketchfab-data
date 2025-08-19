package service

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// LogEntry 日誌條目結構
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Service   string                 `json:"service"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// LogService 日誌服務
type LogService struct {
	conn    net.Conn
	host    string
	port    string
	service string
}

// NewLogService 建立新的日誌服務
func NewLogService(host, port, service string) *LogService {
	return &LogService{
		host:    host,
		port:    port,
		service: service,
	}
}

// Connect 連接到 Logstash
func (ls *LogService) Connect() error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", ls.host, ls.port), 5*time.Second)
	if err != nil {
		return fmt.Errorf("連接到 Logstash 失敗: %v", err)
	}
	ls.conn = conn
	return nil
}

// Close 關閉連接
func (ls *LogService) Close() error {
	if ls.conn != nil {
		return ls.conn.Close()
	}
	return nil
}

// Log 發送日誌到 Logstash
func (ls *LogService) Log(level, message string) error {
	if ls.conn == nil {
		// 如果沒有連接，嘗試重新連接
		if err := ls.Connect(); err != nil {
			// 如果連接失敗，只輸出到標準輸出
			fmt.Printf("[%s] %s: %s\n", level, ls.service, message)
			return nil
		}
	}

	logEntry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Service:   ls.service,
		Type:      "sketchfab",
	}

	data, err := json.Marshal(logEntry)
	if err != nil {
		return fmt.Errorf("序列化日誌失敗: %v", err)
	}

	// 發送到 Logstash
	_, err = ls.conn.Write(append(data, '\n'))
	if err != nil {
		// 如果發送失敗，關閉連接並輸出到標準輸出
		ls.conn.Close()
		ls.conn = nil
		fmt.Printf("[%s] %s: %s\n", level, ls.service, message)
		return nil
	}

	return nil
}

// Info 發送 INFO 級別日誌
func (ls *LogService) Info(message string) error {
	return ls.Log("INFO", message)
}

// Error 發送 ERROR 級別日誌
func (ls *LogService) Error(message string) error {
	return ls.Log("ERROR", message)
}

// Warn 發送 WARN 級別日誌
func (ls *LogService) Warn(message string) error {
	return ls.Log("WARN", message)
}

// Debug 發送 DEBUG 級別日誌
func (ls *LogService) Debug(message string) error {
	return ls.Log("DEBUG", message)
}

// LogWithData 發送包含資料的日誌
func (ls *LogService) LogWithData(level, message string, data map[string]interface{}) error {
	if ls.conn == nil {
		// 如果沒有連接，嘗試重新連接
		if err := ls.Connect(); err != nil {
			// 如果連接失敗，只輸出到標準輸出
			fmt.Printf("[%s] %s: %s\n", level, ls.service, message)
			return nil
		}
	}

	logEntry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Service:   ls.service,
		Type:      "sketchfab",
		Data:      data,
	}

	dataBytes, err := json.Marshal(logEntry)
	if err != nil {
		return fmt.Errorf("序列化日誌失敗: %v", err)
	}

	// 發送到 Logstash
	_, err = ls.conn.Write(append(dataBytes, '\n'))
	if err != nil {
		// 如果發送失敗，關閉連接並輸出到標準輸出
		ls.conn.Close()
		ls.conn = nil
		fmt.Printf("[%s] %s: %s\n", level, ls.service, message)
		return nil
	}

	return nil
}

// LogAPIData 記錄 API 資料
func (ls *LogService) LogAPIData(message string, apiData interface{}) error {
	data := map[string]interface{}{
		"api_data":  apiData,
		"data_type": "api_response",
	}
	return ls.LogWithData("INFO", message, data)
}

// LogModelData 記錄模型資料
func (ls *LogService) LogModelData(message string, modelData interface{}) error {
	data := map[string]interface{}{
		"model_data": modelData,
		"data_type":  "model_info",
	}
	return ls.LogWithData("INFO", message, data)
}
