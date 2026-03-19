package immich

import (
	"fmt"
	"log/slog"
	"path"

	"github.com/google/uuid"
)

func GetAlbums(server ServerConfig) (
	albums []AlbumResponseDto,
	err error,
) {
	if server.DryRun {
		slog.Debug("Dry run: return empty album list")
		return []AlbumResponseDto{}, nil
	}
	return Get[[]AlbumResponseDto](server, "/api/albums")
}

func DeleteEmptyAlbums(server ServerConfig) error {
	if server.DryRun {
		slog.Debug("Dry run: skipping empty album deletion")
		return nil
	}

	albums, err := GetAlbums(server)
	if err != nil {
		return fmt.Errorf("failed to get albums: %w", err)
	}

	slog.Debug("albums", slog.Any("albums", albums))
	for _, album := range albums {
		if album.AssetCount != 0 {
			continue
		}

		slog.Debug("Deleting album",
			slog.String("name", album.AlbumName),
			slog.String("id", album.Id),
		)
		err = DeleteAlbum(server, album.Id)
		if err != nil {
			return fmt.Errorf("failed to delete album '%s': %w",
				album.AlbumName, err)
		}
	}
	return nil
}

func DeleteAlbum(server ServerConfig, id string) error {
	if server.DryRun {
		slog.Debug("Dry run: skipping album deletion",
			slog.String("id", id),
		)
		return nil
	}

	resp, err := DoRequest(
		"DELETE", server, path.Join("api", "albums", id), nil, "")
	if err != nil {
		return err
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status code %d: %s",
			resp.StatusCode,
			resp.Status,
		)
	}

	return nil
}

func CreateAlbum(server ServerConfig, albumName string, assetIds []string) (
	album CreateAlbumDto,
	err error,
) {
	if server.DryRun {
		slog.Debug("Dry run: return fake album",
			slog.String("name", albumName),
			slog.Int("assetCount", len(assetIds)),
		)

		album = CreateAlbumDto{
			ID:        uuid.NewString(),
			AlbumName: albumName,
		}
		return
	}

	return Post[CreateAlbumDto](
		server,
		"/api/albums",
		CreateAlbumRequest{
			AlbumName: albumName,
			AssetIDs:  assetIds,
		},
	)
}
