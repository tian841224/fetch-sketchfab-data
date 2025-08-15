package main

import (
	"fmt"
	"log"

	"fetch-sketchfab-data/internal/api"
	"fetch-sketchfab-data/internal/config"
	"fetch-sketchfab-data/internal/database"
	"fetch-sketchfab-data/internal/service"
)

func main() {
	// 載入設定
	cfg := config.LoadConfig()

	// 建立MongoDB連線
	mongoConfig := &database.MongoDBConfig{
		URI:      cfg.MongoDB.URI,
		Database: cfg.MongoDB.Database,
		Timeout:  cfg.MongoDB.Timeout,
	}

	mongoClient, err := database.NewMongoDBClient(mongoConfig)
	if err != nil {
		log.Fatalf("MongoDB連線失敗: %v", err)
	}
	defer func() {
		if err := mongoClient.Close(); err != nil {
			log.Printf("關閉MongoDB連線時發生錯誤: %v", err)
		}
	}()

	// 建立模型服務
	modelsService := service.NewModelsService(mongoClient)

	// 顯示連線資訊
	connInfo := mongoClient.GetConnectionInfo()
	fmt.Printf("MongoDB連線資訊: %+v\n", connInfo)

	// 建立 Sketchfab API 客戶端
	client := api.NewSketchfabClient()
	response, err := client.GetDownloadableModels()
	if err != nil {
		log.Fatalf("API呼叫失敗: %v", err)
	}

	fmt.Printf("成功取得 %d 個模型資料\n", len(response.Results))

	// 將API回應儲存到資料庫
	fmt.Println("正在將模型資料儲存到資料庫...")
	err = modelsService.ConvertAndSaveModelsResponse(response)
	if err != nil {
		log.Fatalf("儲存模型資料失敗: %v", err)
	}

	fmt.Println("模型資料已成功儲存到資料庫!")

	// 顯示資料庫統計
	totalCount, err := modelsService.GetModelsCount()
	if err != nil {
		log.Printf("取得模型總數失敗: %v", err)
	} else {
		fmt.Printf("資料庫中的模型總數: %d\n", totalCount)
	}

}
