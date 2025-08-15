package service

import (
	"context"
	"fmt"
	"time"

	"fetch-sketchfab-data/internal/database"
	"fetch-sketchfab-data/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ModelsService 處理模型相關的資料庫操作
type ModelsService struct {
	client     *database.MongoDBClient
	collection *mongo.Collection
}

// SketchfabModel 代表Sketchfab模型的資料結構
type SketchfabModel struct {
	ID             string                 `bson:"_id" json:"id"`
	Name           string                 `bson:"name" json:"name"`
	Description    string                 `bson:"description" json:"description"`
	URI            string                 `bson:"uri" json:"uri"`
	User           map[string]interface{} `bson:"user" json:"user"`
	License        map[string]interface{} `bson:"license" json:"license"`
	Tags           []map[string]string    `bson:"tags" json:"tags"`
	Categories     []map[string]string    `bson:"categories" json:"categories"`
	CreatedAt      time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time              `bson:"updated_at" json:"updated_at"`
	FetchedAt      time.Time              `bson:"fetched_at" json:"fetched_at"`
	ViewCount      int                    `bson:"view_count" json:"view_count"`
	LikeCount      int                    `bson:"like_count" json:"like_count"`
	IsDownloadable bool                   `bson:"is_downloadable" json:"is_downloadable"`
	RawData        map[string]interface{} `bson:"raw_data" json:"raw_data"` // 儲存原始API回應
}

// NewModelsService 建立新的模型服務
func NewModelsService(client *database.MongoDBClient) *ModelsService {
	collection := client.GetCollection("models")

	// 建立索引以提升查詢效能
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 建立複合索引
		indexes := []mongo.IndexModel{
			{
				Keys: bson.D{
					{Key: "name", Value: 1},
				},
			},
			{
				Keys: bson.D{
					{Key: "uid", Value: 1},
				},
			},
		}

		collection.Indexes().CreateMany(ctx, indexes)
	}()

	return &ModelsService{
		client:     client,
		collection: collection,
	}
}

// SaveModel 儲存或更新模型
func (s *ModelsService) SaveModel(model *SketchfabModel) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	model.FetchedAt = time.Now()

	// 使用upsert來更新或插入
	filter := bson.M{"_id": model.ID}
	update := bson.M{"$set": model}

	opts := options.Update().SetUpsert(true)

	_, err := s.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("儲存模型失敗: %v", err)
	}

	return nil
}

// SaveModels 批次儲存多個模型
func (s *ModelsService) SaveModels(models []*SketchfabModel) error {
	if len(models) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 準備批次操作
	var operations []mongo.WriteModel

	for _, model := range models {
		model.FetchedAt = time.Now()

		operation := mongo.NewUpdateOneModel()
		operation.SetFilter(bson.M{"_id": model.ID})
		operation.SetUpdate(bson.M{"$set": model})
		operation.SetUpsert(true)

		operations = append(operations, operation)
	}

	// 執行批次寫入
	_, err := s.collection.BulkWrite(ctx, operations)
	if err != nil {
		return fmt.Errorf("批次儲存模型失敗: %v", err)
	}

	return nil
}

// GetModelByID 根據ID取得模型
func (s *ModelsService) GetModelByID(id string) (*SketchfabModel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var model SketchfabModel
	err := s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&model)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("找不到ID為 %s 的模型", id)
		}
		return nil, fmt.Errorf("查詢模型失敗: %v", err)
	}

	return &model, nil
}

// GetModelsCount 取得模型總數
func (s *ModelsService) GetModelsCount() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := s.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("計算模型數量失敗: %v", err)
	}

	return count, nil
}

// UpsertResult 表示 upsert 操作的結果
type UpsertResult struct {
	InsertedCount  int64 `json:"inserted_count"`
	UpdatedCount   int64 `json:"updated_count"`
	UnchangedCount int64 `json:"unchanged_count"`
}

// UpsertModels - 只在資料有變化時才更新
func (s *ModelsService) UpsertModels(models []*SketchfabModel) (*UpsertResult, error) {
	if len(models) == 0 {
		return &UpsertResult{}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result := &UpsertResult{}
	var operations []mongo.WriteModel

	for _, model := range models {
		// 檢查現有資料
		var existingModel SketchfabModel
		err := s.collection.FindOne(ctx, bson.M{"_id": model.ID}).Decode(&existingModel)

		if err == mongo.ErrNoDocuments {
			// 資料不存在，準備插入
			model.FetchedAt = time.Now()

			operation := mongo.NewUpdateOneModel()
			operation.SetFilter(bson.M{"_id": model.ID})
			operation.SetUpdate(bson.M{"$set": model})
			operation.SetUpsert(true)

			operations = append(operations, operation)
			result.InsertedCount++

		} else if err == nil {
			// 資料存在，檢查是否需要更新
			if s.shouldUpdateModel(&existingModel, model) {
				// 保留原始建立時間
				model.CreatedAt = existingModel.CreatedAt
				model.FetchedAt = time.Now()

				operation := mongo.NewUpdateOneModel()
				operation.SetFilter(bson.M{"_id": model.ID})
				operation.SetUpdate(bson.M{"$set": model})

				operations = append(operations, operation)
				result.UpdatedCount++
			} else {
				// 資料沒有變化，只更新取得時間
				operation := mongo.NewUpdateOneModel()
				operation.SetFilter(bson.M{"_id": model.ID})
				operation.SetUpdate(bson.M{"$set": bson.M{"fetched_at": time.Now()}})

				operations = append(operations, operation)
				result.UnchangedCount++
			}
		} else {
			return nil, fmt.Errorf("檢查現有模型失敗: %v", err)
		}
	}

	// 執行批次操作
	if len(operations) > 0 {
		_, err := s.collection.BulkWrite(ctx, operations)
		if err != nil {
			return nil, fmt.Errorf("批次 upsert 失敗: %v", err)
		}
	}

	return result, nil
}

// shouldUpdateModel 判斷模型是否需要更新
func (s *ModelsService) shouldUpdateModel(existing, new *SketchfabModel) bool {
	// 檢查關鍵欄位是否有變化
	if existing.Name != new.Name ||
		existing.Description != new.Description ||
		existing.ViewCount != new.ViewCount ||
		existing.LikeCount != new.LikeCount ||
		existing.IsDownloadable != new.IsDownloadable {
		return true
	}

	// 檢查更新時間
	if !existing.UpdatedAt.Equal(new.UpdatedAt) {
		return true
	}

	// 檢查標籤數量變化
	if len(existing.Tags) != len(new.Tags) {
		return true
	}

	// 檢查分類數量變化
	if len(existing.Categories) != len(new.Categories) {
		return true
	}

	return false
}

// ConvertAndSaveModelsResponse 將API回應轉換為資料庫模型並儲存
func (s *ModelsService) ConvertAndSaveModelsResponse(response *models.ModelsResponse) (*UpsertResult, error) {
	if response == nil || len(response.Results) == 0 {
		return nil, fmt.Errorf("回應為空或沒有模型資料")
	}

	// 轉換API模型為資料庫模型
	var dbModels []*SketchfabModel

	for _, apiModel := range response.Results {
		// 解析時間
		createdAt, _ := time.Parse(time.RFC3339, apiModel.CreatedAt)
		updatedAt, _ := time.Parse(time.RFC3339, apiModel.PublishedAt)

		// 轉換標籤
		tags := make([]map[string]string, len(apiModel.Tags))
		for i, tag := range apiModel.Tags {
			tags[i] = map[string]string{
				"name": tag.Name,
				"slug": tag.Slug,
				"uri":  tag.URI,
			}
		}

		// 轉換分類
		categories := make([]map[string]string, len(apiModel.Categories))
		for i, category := range apiModel.Categories {
			categories[i] = map[string]string{
				"name": category.Name,
			}
		}

		// 轉換使用者資訊
		user := map[string]interface{}{
			"uid":         apiModel.User.UID,
			"username":    apiModel.User.Username,
			"displayName": apiModel.User.DisplayName,
			"profileUrl":  apiModel.User.ProfileURL,
			"account":     apiModel.User.Account,
			"uri":         apiModel.User.URI,
		}

		// 轉換授權資訊
		license := map[string]interface{}{
			"uid":   apiModel.License.UID,
			"label": apiModel.License.Label,
		}

		// 建立資料庫模型
		dbModel := &SketchfabModel{
			ID:             apiModel.UID,
			Name:           apiModel.Name,
			Description:    apiModel.Description,
			URI:            apiModel.URI,
			User:           user,
			License:        license,
			Tags:           tags,
			Categories:     categories,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
			ViewCount:      apiModel.ViewCount,
			LikeCount:      apiModel.LikeCount,
			IsDownloadable: apiModel.IsDownloadable,
			RawData: map[string]interface{}{
				"thumbnails":      apiModel.Thumbnails,
				"archives":        apiModel.Archives,
				"viewerUrl":       apiModel.ViewerURL,
				"embedUrl":        apiModel.EmbedURL,
				"commentCount":    apiModel.CommentCount,
				"animationCount":  apiModel.AnimationCount,
				"faceCount":       apiModel.FaceCount,
				"vertexCount":     apiModel.VertexCount,
				"soundCount":      apiModel.SoundCount,
				"isAgeRestricted": apiModel.IsAgeRestricted,
				"isProtected":     apiModel.IsProtected,
				"price":           apiModel.Price,
				"staffPickedAt":   apiModel.StaffPickedAt,
			},
		}

		dbModels = append(dbModels, dbModel)
	}

	return s.UpsertModels(dbModels)
}
