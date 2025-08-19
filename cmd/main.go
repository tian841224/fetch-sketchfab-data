package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"fetch-sketchfab-data/internal/api"
	"fetch-sketchfab-data/internal/config"
	"fetch-sketchfab-data/internal/database"
	"fetch-sketchfab-data/internal/scheduler"
	"fetch-sketchfab-data/internal/service"
)

func main() {
	// 命令列參數
	var (
		mode         = flag.String("mode", "once", "執行模式: once(單次執行) 或 schedule(排程執行)")
		scheduleTime = flag.String("time", "09:00", "排程執行時間 (格式: HH:MM, 24小時制)")
	)
	flag.Parse()

	// 顯示使用說明
	if *mode != "once" && *mode != "schedule" {
		fmt.Println("使用方式:")
		fmt.Println("  單次執行: go run cmd/main.go -mode=once")
		fmt.Println("  排程執行: go run cmd/main.go -mode=schedule -time=09:00")
		os.Exit(1)
	}
	// 載入設定
	cfg := config.LoadConfig()

	// 建立MongoDB連線
	mongoConfig := &database.MongoDBConfig{
		URI:      cfg.MongoDB.URI,
		Database: cfg.MongoDB.Database,
		Timeout:  cfg.MongoDB.Timeout,
	}

	// 建立 MongoDB Client端
	mongoClient, err := database.NewMongoDBClient(mongoConfig)
	if err != nil {
		log.Fatalf("MongoDB連線失敗: %v", err)
	}
	defer func() {
		if err := mongoClient.Close(); err != nil {
			log.Printf("關閉MongoDB連線時發生錯誤: %v", err)
		}
	}()

	// 建立日誌服務
	logService := service.NewLogService(cfg.Logstash.Host, cfg.Logstash.Port, "sketchfab-fetcher")
	defer logService.Close()

	// 輸出啟動訊息到標準輸出
	fmt.Printf("⏰ 啟動每日排程模式，執行時間: %s\n", *scheduleTime)

	// 建立模型服務
	modelsService := service.NewModelsService(mongoClient)

	// 建立 Sketchfab API 客戶端
	client := api.NewSketchfabClient()

	// 根據模式執行
	switch *mode {
	case "once":
		logService.Info("🔧 執行單次同步...")
		err = runOnce(client, modelsService, logService)
		if err != nil {
			logService.Error(fmt.Sprintf("單次執行失敗: %v", err))
			log.Fatalf("單次執行失敗: %v", err)
		}
		logService.Info("✅ 單次執行完成!")

	case "schedule":
		logService.Info(fmt.Sprintf("⏰ 啟動每日排程模式，執行時間: %s", *scheduleTime))
		err = runScheduler(client, modelsService, logService, *scheduleTime)
		if err != nil {
			logService.Error(fmt.Sprintf("排程器執行失敗: %v", err))
			log.Fatalf("排程器執行失敗: %v", err)
		}
	}
}

// runOnce 執行單次同步
func runOnce(client *api.SketchfabClient, modelsService *service.ModelsService, logService *service.LogService) error {
	response, err := client.GetDownloadableModels()
	if err != nil {
		logService.Error(fmt.Sprintf("API呼叫失敗: %v", err))
		return fmt.Errorf("API呼叫失敗: %v", err)
	}

	logService.Info(fmt.Sprintf("📥 成功取得 %d 個模型資料", len(response.Results)))

	// 將API回應儲存到資料庫
	logService.Info("正在將模型資料儲存到資料庫...")
	upsertResult, err := modelsService.ConvertAndSaveModelsResponse(response)
	if err != nil {
		logService.Error(fmt.Sprintf("儲存模型資料失敗: %v", err))
		return fmt.Errorf("儲存模型資料失敗: %v", err)
	}

	// 顯示 upsert 統計結果
	logService.Info(fmt.Sprintf("📊 處理統計: 新增=%d, 更新=%d, 無變化=%d",
		upsertResult.InsertedCount, upsertResult.UpdatedCount, upsertResult.UnchangedCount))

	// 顯示資料庫統計
	totalCount, err := modelsService.GetModelsCount()
	if err != nil {
		logService.Error(fmt.Sprintf("取得模型總數失敗: %v", err))
	} else {
		logService.Info(fmt.Sprintf("💾 資料庫中的模型總數: %d", totalCount))
	}

	return nil
}

// runScheduler 執行排程器模式
func runScheduler(client *api.SketchfabClient, modelsService *service.ModelsService, logService *service.LogService, scheduleTime string) error {
	// 建立每日排程器
	dailyScheduler := scheduler.NewDailyScheduler(client, modelsService, logService, scheduleTime)

	// 建立 context 和信號處理
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 設定信號處理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 在 goroutine 中啟動排程器
	errChan := make(chan error, 1)
	go func() {
		errChan <- dailyScheduler.Start(ctx)
	}()

	// 等待信號或錯誤
	select {
	case <-sigChan:
		logService.Info("收到停止信號，正在關閉...")
		dailyScheduler.Stop()
		cancel()
		return nil
	case err := <-errChan:
		logService.Error(fmt.Sprintf("排程器錯誤: %v", err))
		return err
	}
}
