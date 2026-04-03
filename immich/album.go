package immich

import (
	"context"
	"fmt"
	"log/slog"
	"path"

	"github.com/google/uuid"
)

func GetAlbums(ctx context.Context, server ServerConfig) (
	albums []AlbumResponseDto,
	err error,
) {
	if ctx.Err() != nil {
		err = fmt.Errorf("context error: %w", ctx.Err())
		return
	}

	if server.DryRun {
		slog.Debug("Dry run: return empty album list")
		return []AlbumResponseDto{}, nil
	}
	return Get[[]AlbumResponseDto](ctx, server, "/api/albums")
}

func GetAlbum(ctx context.Context, server ServerConfig, id string) (
	album AlbumResponseDto,
	err error,
) {
	if ctx.Err() != nil {
		err = fmt.Errorf("context error: %w", ctx.Err())
		return
	}

	if server.DryRun {
		slog.Debug("Dry run: return empty album list")
		return AlbumResponseDto{}, nil
	}
	return Get[AlbumResponseDto](ctx, server, path.Join("/api/albums", id))
}

func DeleteEmptyAlbums(ctx context.Context, server ServerConfig) error {
	if ctx.Err() != nil {
		return fmt.Errorf("context error: %w", ctx.Err())
	}

	if server.DryRun {
		slog.Debug("Dry run: skipping empty album deletion")
		return nil
	}

	albums, err := GetAlbums(ctx, server)
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
			slog.String("id", album.ID),
		)
		err = DeleteAlbum(ctx, server, album.ID)
		if err != nil {
			return fmt.Errorf("failed to delete album '%s': %w",
				album.AlbumName, err)
		}
	}
	return nil
}

func DeleteAlbum(ctx context.Context, server ServerConfig, id string) error {
	if ctx.Err() != nil {
		return fmt.Errorf("context error: %w", ctx.Err())
	}

	if server.DryRun {
		slog.Debug("Dry run: skipping album deletion",
			slog.String("id", id),
		)
		return nil
	}

	resp, err := DoRequest(
		ctx, "DELETE", server, path.Join("api", "albums", id), nil, "")
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

func CreateAlbum(ctx context.Context, server ServerConfig, albumName string, assetIds []string) (
	album CreateAlbumDto,
	err error,
) {
	if ctx.Err() != nil {
		err = fmt.Errorf("context error: %w", ctx.Err())
		return
	}

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
		ctx,
		server,
		"/api/albums",
		CreateAlbumRequest{
			AlbumName: albumName,
			AssetIDs:  assetIds,
		},
	)
}

func AddAssetsToAlbum(ctx context.Context, server ServerConfig, albumIds []string, assetIds []string) error {
	if ctx.Err() != nil {
		return fmt.Errorf("context error: %w", ctx.Err())
	}

	if server.DryRun {
		slog.Debug("Dry run: skipping adding assets to album",
			slog.Any("albumIds", albumIds),
			slog.Int("assetCount", len(assetIds)),
		)
		return nil
	}

	resp, err := Post[AddAssetsToAlbumResponse](
		ctx,
		server,
		path.Join("api", "albums", "assets"),
		AddAssetsToAlbumRequest{
			AlbumIDS: albumIds,
			AssetIDs: assetIds,
		},
	)

	if err != nil {
		return fmt.Errorf("unable to add assets to albums: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("server error: %s", resp.Error)
	}
	return err
}
