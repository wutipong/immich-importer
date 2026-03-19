package directory

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"slices"

	"github.com/wutipong/immich-importer/immich"
)

func Process(url *url.URL, apiKey string, sourceDir string, force bool) error {
	var albums []immich.AlbumResponseDto
	var err error

	if !force {
		albums, err = immich.GetAlbums(url, apiKey)
		if err != nil {
			err = fmt.Errorf("unable to retrieved existing albums. => %w", err)
			return err
		}
		slog.Debug("albums", slog.Any("existing albums", albums))
	}
	err = filepath.WalkDir(sourceDir,
		func(path string, d os.DirEntry, err error,
		) error {
			if !d.IsDir() {
				return nil
			}

			if err != nil {
				slog.Warn(
					"failed to access path. skipping.",
					slog.String("path", path),
					slog.String("error", err.Error()),
				)
				return nil
			}

			entries, err := os.ReadDir(path)
			if err != nil {
				slog.Warn("failed to read directory", slog.String("path", path))
				return nil
			}
			files := slices.DeleteFunc(
				entries,
				func(d os.DirEntry) bool {
					if d.IsDir() {
						return true
					}
					if !immich.IsMediaFile(d.Name()) {
						return true
					}
					return false
				},
			)

			if len(files) == 0 {
				slog.Info(
					"skipping directory without media files.",
					slog.String("path", path),
				)
				return nil
			}

			archiveFilePath, err := filepath.Rel(sourceDir, path)
			if err != nil {
				slog.Error(
					"failed to determine album name.",
					slog.String("error", err.Error()),
				)
				return nil
			}

			slog.Debug("album name", slog.String("album_name", archiveFilePath))

			if slices.ContainsFunc(albums, func(a immich.AlbumResponseDto) bool {
				return a.AlbumName == archiveFilePath
			}) {
				slog.Warn("album already exists. skipping.", slog.String("album_name", archiveFilePath))
				return nil
			}

			archiveFile, err := os.Open(path)
			if err != nil {
				slog.Warn("failed to open archive", slog.String("file", archiveFilePath))
				return nil
			}
			defer archiveFile.Close()

			assetIds := make([]string, 0)

			for _, file := range files {
				slog.Info(
					"creating asset",
					slog.String("album", archiveFilePath),
					slog.String("entry", file.Name()),
				)
				info, err := file.Info()
				if err != nil {
					slog.Error("failed to get file info",
						slog.String("album", archiveFilePath),
						slog.String("entry", file.Name()),
						slog.String("error", err.Error()),
					)
					return nil
				}

				reader, err := os.Open(filepath.Join(path, file.Name()))
				if err != nil {
					slog.Error("failed to open file",
						slog.String("album", archiveFilePath),
						slog.String("entry", reader.Name()),
						slog.String("error", err.Error()),
					)
					return nil
				}

				defer reader.Close()

				asset, err := immich.PostAsset(
					archiveFilePath, reader.Name(), reader, info.ModTime(), url, apiKey,
				)
				if err != nil {
					slog.Error("failed to upload asset",
						slog.String("album", archiveFilePath),
						slog.String("entry", reader.Name()),
						slog.String("error", err.Error()),
					)
					continue
				}

				slog.Info("uploaded asset", slog.Any("asset", asset))

				assetIds = append(assetIds, asset.ID)
			}

			if len(assetIds) == 0 {
				slog.Warn(
					"no asset uploaded. skipping album creation.",
					slog.String("album_name", archiveFilePath),
				)
				return nil
			}

			slog.Info("creating album", slog.String("name", archiveFilePath))

			createdAlbum, err := immich.CreateAlbum(
				url, apiKey, archiveFilePath, assetIds,
			)
			if err != nil {
				slog.Error("failed to create album", slog.String("error", err.Error()))
				return nil
			}

			slog.Info("created album", slog.Any("album", createdAlbum))

			return nil
		})

	if err != nil {
		slog.Error("fails creating albums", slog.Any("error", err))
	}
	return nil
}
