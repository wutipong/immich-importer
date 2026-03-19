package directory

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/wutipong/immich-importer/immich"
)

func Process(
	server immich.ServerConfig,
	sourceDir string,
	path string,
) (
	assetIds []string,
	err error,
) {
	entries, err := os.ReadDir(path)
	if err != nil {
		err = fmt.Errorf("failed to read directory. => %w", err)
		return
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

	archiveFilePath, err := filepath.Rel(sourceDir, path)
	if err != nil {
		err = fmt.Errorf(
			"failed to determine album name. => %w",
			err,
		)
		return
	}

	for _, file := range files {
		slog.Info(
			"creating asset",
			slog.String("album", archiveFilePath),
			slog.String("entry", file.Name()),
		)

		info, e := file.Info()
		if e != nil {
			err = fmt.Errorf(
				"Unable to read image file propery: %s. => %w",
				file.Name(),
				e,
			)
			return
		}

		reader, e := os.Open(filepath.Join(path, file.Name()))
		if err != nil {
			err = fmt.Errorf(
				"failed to open image file %s. => %w",
				file.Name(),
				e,
			)
			return
		}

		defer reader.Close()

		asset, err := immich.PostAsset(
			server,
			archiveFilePath,
			file.Name(),
			reader,
			info.ModTime(),
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

	return
}
