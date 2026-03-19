package immich

import "time"

type AlbumResponseDto struct {
	AlbumName  string `json:"albumName"`
	Id         string `json:"id"`
	AssetCount int64  `json:"assetCount"`
}

type CreateAlbumRequest struct {
	AlbumName string   `json:"albumName"`
	AssetIDs  []string `json:"assetIds"`
}

type CreateAlbumDto struct {
	AlbumName string `json:"albumName"`
	ID        string `json:"id"`
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
