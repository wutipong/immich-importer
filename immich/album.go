package immich

import (
	"fmt"
	"log/slog"
	"net/url"
)

func GetAlbums(url *url.URL, apiKey string) (albums []AlbumResponseDto, err error) {
	return Get[[]AlbumResponseDto](url.JoinPath("/api/albums"), apiKey)
}

func DeleteEmptyAlbums(url *url.URL, apiKey string) error {
	albums, err := GetAlbums(url, apiKey)
	if err != nil {
		return fmt.Errorf("failed to get albums: %w", err)
	}

	slog.Debug("albums", slog.Any("albums", albums))
	for _, album := range albums {
		if album.AssetCount != 0 {
			continue
		}

		slog.Debug("Deleting album", slog.String("name", album.AlbumName), slog.String("id", album.Id))
		err = DeleteAlbum(album.Id, url, apiKey)
		if err != nil {
			return fmt.Errorf("failed to delete album '%s': %w", album.AlbumName, err)
		}
	}
	return nil
}

func DeleteAlbum(id string, url *url.URL, apiKey string) error {
	resp, err := DoRequest("DELETE", url.JoinPath("api", "albums", id), nil, "", apiKey)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, resp.Status)
	}

	return nil
}

func CreateAlbum(url *url.URL, apiKey string, albumName string, assetIds []string) (CreateAlbumDto, error) {
	return Post[CreateAlbumDto](
		url.JoinPath("/api/albums"),
		CreateAlbumRequest{
			AlbumName: albumName,
			AssetIDs:  assetIds,
		},
		apiKey,
	)
}
