package main

import (
	"fmt"
	"log"

	"fetch-sketchfab-data/internal/api"
)

func main() {
	// 建立 Sketchfab API 客戶端
	client := api.NewSketchfabClient()
	response, err := client.GetDownloadableModels()
	if err != nil {
		log.Fatalf("API呼叫失敗: %v", err)
	}
	fmt.Println(response)
}
