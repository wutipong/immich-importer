package merge

import (
	"fmt"
	"log/slog"
	"regexp"
	"slices"

	"github.com/wutipong/immich-importer/immich"
)

func Process(server immich.ServerConfig, album string, pattern string) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	albums, err := immich.GetAlbums(server)
	if err != nil {
		return fmt.Errorf("unable to retrive albums: %w", err)
	}

	albums = slices.DeleteFunc(
		albums,
		func(album immich.AlbumResponseDto) bool {
			return !regex.MatchString(album.AlbumName)
		},
	)

	albumSets := make(map[string]bool)
	for _, album := range albums {
		for _, asset := range album.Assets {
			albumSets[asset.ID] = true
		}
	}

	assetIds := make([]string, 0)
	for key := range albumSets {
		assetIds = append(assetIds, key)
	}

	slog.Info("creating album", slog.String("name", album))
	createdAlbum, err := immich.CreateAlbum(
		server, album, assetIds,
	)

	if err != nil {
		return fmt.Errorf("faield to create album: %w", err)
	}

	slog.Info(
		"created album",
		slog.String("id", createdAlbum.ID),
		slog.String("name", createdAlbum.AlbumName),
	)

	return nil
}
