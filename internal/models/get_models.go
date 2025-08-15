package models

// Cursors 代表分頁游標資訊
type Cursors struct {
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
}

// ThumbnailImage 代表縮圖圖片資訊
type ThumbnailImage struct {
	UID    string `json:"uid"`
	Size   int    `json:"size"`
	Width  int    `json:"width"`
	URL    string `json:"url"`
	Height int    `json:"height"`
}

// Thumbnails 代表縮圖集合
type Thumbnails struct {
	Images []ThumbnailImage `json:"images"`
}

// Tag 代表標籤資訊
type Tag struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	URI  string `json:"uri"`
}

// Category 代表分類資訊
type Category struct {
	Name string `json:"name"`
}

// License 代表授權資訊
type License struct {
	UID   string `json:"uid"`
	Label string `json:"label"`
}

// Avatar 代表使用者頭像資訊
type Avatar struct {
	URI    string           `json:"uri"`
	Images []ThumbnailImage `json:"images"`
}

// User 代表使用者資訊
type User struct {
	UID         string  `json:"uid"`
	Username    string  `json:"username"`
	DisplayName string  `json:"displayName"`
	ProfileURL  *string `json:"profileUrl"`
	Account     string  `json:"account"`
	Avatar      Avatar  `json:"avatar"`
	URI         string  `json:"uri"`
}

// Archive 代表檔案封存資訊
type Archive struct {
	TextureCount         *int   `json:"textureCount"`
	Size                 int    `json:"size"`
	Type                 string `json:"type"`
	TextureMaxResolution *int   `json:"textureMaxResolution"`
	FaceCount            *int   `json:"faceCount"`
	VertexCount          *int   `json:"vertexCount"`
}

// Archives 代表各種格式的檔案封存
type Archives struct {
	GLB    *Archive `json:"glb,omitempty"`
	GLTF   *Archive `json:"gltf,omitempty"`
	GLTFAR *Archive `json:"gltf-ar,omitempty"`
	Source *Archive `json:"source,omitempty"`
	USDZ   *Archive `json:"usdz,omitempty"`
}

// Model 代表 3D 模型資訊
type Model struct {
	URI             string     `json:"uri"`
	UID             string     `json:"uid"`
	Name            string     `json:"name"`
	StaffPickedAt   *string    `json:"staffpickedAt"`
	ViewCount       int        `json:"viewCount"`
	LikeCount       int        `json:"likeCount"`
	AnimationCount  int        `json:"animationCount"`
	ViewerURL       string     `json:"viewerUrl"`
	EmbedURL        string     `json:"embedUrl"`
	CommentCount    int        `json:"commentCount"`
	IsDownloadable  bool       `json:"isDownloadable"`
	PublishedAt     string     `json:"publishedAt"`
	Tags            []Tag      `json:"tags"`
	Categories      []Category `json:"categories"`
	Thumbnails      Thumbnails `json:"thumbnails"`
	User            User       `json:"user"`
	Description     string     `json:"description"`
	FaceCount       int        `json:"faceCount"`
	CreatedAt       string     `json:"createdAt"`
	VertexCount     int        `json:"vertexCount"`
	IsAgeRestricted bool       `json:"isAgeRestricted"`
	SoundCount      int        `json:"soundCount"`
	IsProtected     bool       `json:"isProtected"`
	License         License    `json:"license"`
	Price           *float64   `json:"price"`
	Archives        Archives   `json:"archives"`
}

// ModelsResponse 代表模型列表的 API 回應
type ModelsResponse struct {
	Cursors  Cursors `json:"cursors"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []Model `json:"results"`
}

// GetModelsParams 代表獲取模型列表的參數
type GetModelsParams struct {
	Downloadable     bool    `json:"downloadable,omitempty"`
	ArchivesFlavours bool    `json:"archives_flavours,omitempty"`
	Cursor           *string `json:"cursor,omitempty"`
	Count            *int    `json:"count,omitempty"`
	Sort             *string `json:"sort,omitempty"`
	Categories       *string `json:"categories,omitempty"`
	Tags             *string `json:"tags,omitempty"`
	Search           *string `json:"search,omitempty"`
}
