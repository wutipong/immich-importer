package immich

import "time"

type AlbumResponseDto struct {
	AlbumName  string             `json:"albumName"`
	ID         string             `json:"id"`
	AssetCount int64              `json:"assetCount"`
	Assets     []AssetResponseDto `json:"assets"`
}

type CreateAlbumRequest struct {
	AlbumName string   `json:"albumName"`
	AssetIDs  []string `json:"assetIds"`
}

type CreateAlbumDto struct {
	AlbumName string `json:"albumName"`
	ID        string `json:"id"`
}

type AddAssetsToAlbumRequest struct {
	AlbumIDS []string `json:"albumIds"`
	AssetIDs []string `json:"assetIds"`
}

type AddAssetsToAlbumResponse struct {
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

type AssetMediaRequest struct {
	DeviceAssetId  string    `json:"deviceAssetId"`
	DeviceId       string    `json:"deviceId"`
	FileCreatedAt  time.Time `json:"fileCreatedAt"`
	FileModifiedAt time.Time `json:"fileModifiedAt"`
	Filename       string    `json:"filename"`
}

type AssetMediaResponseDto struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type AssetResponseDto struct {
	ID string `json:"id"`
}
