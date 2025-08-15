## 🚀 功能概述

此程式支援兩種執行模式：
1. **單次執行**：立即執行一次資料同步
2. **每日排程**：在指定時間每天自動執行資料同步

## 📋 使用方式

### 1.直接執行

### 單次執行
```bash
# 立即執行一次同步
go run cmd/main.go

```

### 每日排程執行
```bash
# 設定每天 09:00 執行
go run cmd/main.go -mode=schedule -time=09:00

# 設定每天 14:30 執行
go run cmd/main.go -mode=schedule -time=14:30

# 設定每天 23:00 執行
go run cmd/main.go -mode=schedule -time=23:00
```

## ⚙️ 參數說明

| 參數 | 預設值 | 說明 |
|------|--------|------|
| `-mode` | `once` | 執行模式：`once`(單次) 或 `schedule`(排程) |
| `-time` | `09:00` | 排程執行時間，格式為 `HH:MM` (24小時制) |

## 🔧 編譯執行檔

```bash
# 編譯為執行檔
go build -o fetch-sketchfab cmd/main.go

# 單次執行
./fetch-sketchfab -mode=once

# 排程執行
./fetch-sketchfab -mode=schedule -time=09:00
```

## 🖥️ 執行範例

### 單次執行輸出範例
```
🔧 執行單次同步...
📥 成功取得 24 個模型資料
正在將模型資料儲存到資料庫...
📊 處理統計:
   ✅ 新增: 5 個模型
   🔄 更新: 3 個模型
   ⏭️  無變化: 16 個模型
💾 資料庫中的模型總數: 120
✅ 單次執行完成!
```

### 排程執行輸出範例
```
⏰ 啟動每日排程模式，執行時間: 09:00
🕒 每日排程器已啟動，執行時間: 09:00
執行初始資料同步...
📥 成功取得 24 個模型資料
⏰ 下次執行時間: 2024-01-15 09:00:00 (等待 8h45m30s)
```

## 🛑 停止排程器

在排程模式下，可以使用以下方式停止程式：
- 按 `Ctrl+C`



### 使用docker 執行


### 啟動服務 包含 MongoDB 管理介面

```bash
docker-compose --profile admin up -d
```

## 執行模式

### 排程執行模式

若要改為排程模式，編輯 `docker-compose.yml` 檔案中的 `fetch-sketchfab` 服務：

```yaml
# 註解掉單次模式
# command: ["-mode=once"]

# 取消註解排程模式
command: ["-mode=schedule", "-time=09:00"]
```

然後重新啟動服務：

```bash
docker-compose up -d fetch-sketchfab
```

### 手動執行不同模式

```bash
# 單次執行
docker-compose run --rm fetch-sketchfab -mode=once

# 排程執行（每天上午 9:00）
docker-compose run --rm fetch-sketchfab -mode=schedule -time=09:00

# 排程執行（每天下午 2:30）
docker-compose run --rm fetch-sketchfab -mode=schedule -time=14:30
```

## 服務說明

### MongoDB 服務

- **容器名稱**: `sketchfab-mongodb`
- **連接埠**: `27017`
- **資料持久化**: 使用 Docker Volume 儲存資料
- **健康檢查**: 內建健康檢查機制

### Sketchfab 擷取服務

- **容器名稱**: `sketchfab-fetcher`
- **相依性**: 等待 MongoDB 健康檢查通過後才啟動
- **日誌**: 儲存在 `./logs` 目錄

### MongoDB Express

- **容器名稱**: `sketchfab-mongo-express`
- **連接埠**: `8081`
- **存取網址**: http://localhost:8081
- **啟用方式**: 使用 `--profile admin` 參數
