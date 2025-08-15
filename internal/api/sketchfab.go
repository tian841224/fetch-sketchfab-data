package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"fetch-sketchfab-data/internal/models"
)

type SketchfabClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewSketchfabClient() *SketchfabClient {
	return &SketchfabClient{
		BaseURL: "https://api.sketchfab.com/v3",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetModels API
func (c *SketchfabClient) GetModels(params *models.GetModelsParams) (*models.ModelsResponse, error) {
	// 建立 URL
	apiURL, err := url.Parse(fmt.Sprintf("%s/models", c.BaseURL))
	if err != nil {
		return nil, fmt.Errorf("無法解析 URL: %w", err)
	}

	// 設定查詢參數
	query := apiURL.Query()

	if params != nil {
		if params.Downloadable {
			query.Set("downloadable", "true")
		}

		if !params.ArchivesFlavours {
			query.Set("archives_flavours", "false")
		}

		if params.Cursor != nil {
			query.Set("cursor", *params.Cursor)
		}

		if params.Count != nil {
			query.Set("count", fmt.Sprintf("%d", *params.Count))
		}

		if params.Sort != nil {
			query.Set("sort", *params.Sort)
		}

		if params.Categories != nil {
			query.Set("categories", *params.Categories)
		}

		if params.Tags != nil {
			query.Set("tags", *params.Tags)
		}

		if params.Search != nil {
			query.Set("search", *params.Search)
		}
	}

	apiURL.RawQuery = query.Encode()

	// 建立 HTTP 請求
	req, err := http.NewRequest("GET", apiURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("無法建立 HTTP 請求: %w", err)
	}

	// 設定請求標頭
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "fetch-sketchfab-data/1.0")

	// 發送請求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP 請求失敗: %w", err)
	}
	defer resp.Body.Close()

	// 檢查回應狀態碼
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 請求失敗，狀態碼: %d, 回應: %s", resp.StatusCode, string(body))
	}

	// 讀取回應內容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("無法讀取回應內容: %w", err)
	}

	// 解析 JSON 回應
	var modelsResponse models.ModelsResponse
	if err := json.Unmarshal(body, &modelsResponse); err != nil {
		return nil, fmt.Errorf("無法解析 JSON 回應: %w", err)
	}

	return &modelsResponse, nil
}

// GetDownloadableModels API
func (c *SketchfabClient) GetDownloadableModels() (*models.ModelsResponse, error) {
	params := &models.GetModelsParams{
		Downloadable:     true,
		ArchivesFlavours: false,
	}

	return c.GetModels(params)
}
