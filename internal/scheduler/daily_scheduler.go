package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"fetch-sketchfab-data/internal/api"
	"fetch-sketchfab-data/internal/service"
)

// DailyScheduler 每日排程器
type DailyScheduler struct {
	apiClient     *api.SketchfabClient
	modelsService *service.ModelsService
	scheduleTime  string // 格式: "15:04" (24小時制)
	stopChan      chan struct{}
}

// NewDailyScheduler 建立新的每日排程器
func NewDailyScheduler(apiClient *api.SketchfabClient, modelsService *service.ModelsService, scheduleTime string) *DailyScheduler {
	return &DailyScheduler{
		apiClient:     apiClient,
		modelsService: modelsService,
		scheduleTime:  scheduleTime,
		stopChan:      make(chan struct{}),
	}
}

// Start 啟動每日排程器
func (s *DailyScheduler) Start(ctx context.Context) error {
	log.Printf("🕒 每日排程器已啟動，執行時間: %s", s.scheduleTime)

	// 立即執行一次（可選）
	log.Println("執行初始資料同步...")
	if err := s.fetchAndSaveData(); err != nil {
		log.Printf("初始資料同步失敗: %v", err)
	}

	for {
		// 計算下次執行時間
		nextRun := s.calculateNextRunTime()
		waitDuration := time.Until(nextRun)

		log.Printf("⏰ 下次執行時間: %s (等待 %v)", nextRun.Format("2006-01-02 15:04:05"), waitDuration)

		select {
		case <-ctx.Done():
			log.Println("接收到停止信號，正在關閉每日排程器...")
			return ctx.Err()
		case <-s.stopChan:
			log.Println("每日排程器已停止")
			return nil
		case <-time.After(waitDuration):
			log.Println("🚀 開始執行每日任務...")
			if err := s.fetchAndSaveData(); err != nil {
				log.Printf("❌ 每日任務執行失敗: %v", err)
			} else {
				log.Println("✅ 每日任務執行完成")
			}
		}
	}
}

// Stop 停止每日排程器
func (s *DailyScheduler) Stop() {
	log.Println("正在停止每日排程器...")
	close(s.stopChan)
}

// calculateNextRunTime 計算下次執行時間
func (s *DailyScheduler) calculateNextRunTime() time.Time {
	now := time.Now()

	// 解析設定的時間
	targetTime, err := time.Parse("15:04", s.scheduleTime)
	if err != nil {
		log.Printf("時間格式錯誤，使用預設時間 09:00: %v", err)
		targetTime, _ = time.Parse("15:04", "09:00")
	}

	// 計算今天的執行時間
	today := time.Date(now.Year(), now.Month(), now.Day(),
		targetTime.Hour(), targetTime.Minute(), 0, 0, now.Location())

	// 如果今天的時間已經過了，就安排到明天
	if today.Before(now) || today.Equal(now) {
		today = today.Add(24 * time.Hour)
	}

	return today
}

// fetchAndSaveData 取得並儲存資料
func (s *DailyScheduler) fetchAndSaveData() error {
	startTime := time.Now()

	// 呼叫 API
	response, err := s.apiClient.GetDownloadableModels()
	if err != nil {
		return fmt.Errorf("API呼叫失敗: %v", err)
	}

	log.Printf("📥 成功取得 %d 個模型資料", len(response.Results))

	// 儲存到資料庫
	upsertResult, err := s.modelsService.ConvertAndSaveModelsResponse(response)
	if err != nil {
		return fmt.Errorf("儲存模型資料失敗: %v", err)
	}

	// 記錄統計資訊
	duration := time.Since(startTime)
	log.Printf("⏱️  任務完成 (耗時: %v)", duration)
	log.Printf("📊 處理統計:")
	log.Printf("   ✅ 新增: %d 個模型", upsertResult.InsertedCount)
	log.Printf("   🔄 更新: %d 個模型", upsertResult.UpdatedCount)
	log.Printf("   ⏭️  無變化: %d 個模型", upsertResult.UnchangedCount)

	// 顯示資料庫總數
	totalCount, err := s.modelsService.GetModelsCount()
	if err == nil {
		log.Printf("💾 資料庫中的模型總數: %d", totalCount)
	}

	return nil
}

// RunOnce 執行一次任務（用於手動觸發或測試）
func (s *DailyScheduler) RunOnce() error {
	log.Println("🔧 執行單次任務...")
	return s.fetchAndSaveData()
}
