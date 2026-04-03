package merge

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"slices"

	"github.com/wutipong/immich-importer/immich"
)

func Process(
	ctx context.Context,
	server immich.ServerConfig,
	album string,
	pattern string,
	isDeletingAlbums bool,
) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	albums, err := immich.GetAlbums(ctx, server)
	if err != nil {
		return fmt.Errorf("unable to retrive albums: %w", err)
	}

	albums = slices.DeleteFunc(
		albums,
		func(album immich.AlbumResponseDto) bool {
			slog.Debug("filtering album", slog.String("album", album.AlbumName), slog.String("id", album.ID))
			return !regex.MatchString(album.AlbumName)
		},
	)

	slog.Debug("merging albums", slog.Any("albums", albums))

	assetIdSet := make(map[string]bool)
	for _, album := range albums {

		details, err := immich.GetAlbum(ctx, server, album.ID)
		if err != nil {
			slog.Error("unable to retrieve album information",
				slog.String("error", err.Error()))
		}
		for _, asset := range details.Assets {
			assetIdSet[asset.ID] = true
		}
	}

	assetIds := make([]string, 0)
	for key := range assetIdSet {
		assetIds = append(assetIds, key)
	}

	slog.Info("creating album", slog.String("name", album))
	createdAlbum, err := immich.CreateAlbum(
		ctx, server, album, assetIds,
	)

	if err != nil {
		return fmt.Errorf("faield to create album: %w", err)
	}

	if isDeletingAlbums {
		for _, album := range albums {
			err := immich.DeleteAlbum(ctx, server, album.ID)
			if err != nil {
				slog.Warn("unable to delete album", slog.String("id", err.Error()))
			}
		}
	}

	slog.Info(
		"created album",
		slog.String("id", createdAlbum.ID),
		slog.String("name", createdAlbum.AlbumName),
	)

	return nil
}
