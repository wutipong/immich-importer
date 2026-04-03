package backfill

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/wutipong/immich-importer/config"
	"github.com/wutipong/immich-importer/immich"
	"github.com/wutipong/immich-importer/logging"
)

func backfillDirectory(
	ctx context.Context,
	profile, displayLogLevel, fileLogLevel string,
	sourceDir, inputDir string, dryRun bool,
) error {
	err := logging.Setup(profile, displayLogLevel, true, fileLogLevel)
	if err != nil {
		return fmt.Errorf("unable to setup log: %w", err)
	}
	defer logging.CleanUp()

	c, err := config.LoadConfig(profile)
	if err != nil {
		return fmt.Errorf(
			"unable to load configuration. please run 'immich-importer setup' first: %w",
			err,
		)
	}

	slog.Info("Immich instance",
		slog.String("url", c.ImmichURL),
		slog.String("api_key",
			strings.Repeat("*", len(c.ImmichAPIKey)),
		),
	)

	url, err := url.Parse(c.ImmichURL)
	if err != nil {
		return fmt.Errorf("invalid immich url: %w", err)
	}

	server := immich.ServerConfig{
		URL:    url,
		APIKey: c.ImmichAPIKey,
		DryRun: dryRun,
	}

	slog.Debug("processing directory",
		slog.String("sourceDir", sourceDir),
		slog.String("path", inputDir),
	)
	entries, err := os.ReadDir(filepath.Join(sourceDir, inputDir))
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	slog.Debug("directory entries", slog.Any("size", len(entries)))

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

	slog.Debug("media files", slog.Any("size", len(files)))

	for _, file := range files {
		slog.Info(
			"creating asset",
			slog.String("path", inputDir),
			slog.String("entry", file.Name()),
		)

		info, e := file.Info()
		if e != nil {
			return fmt.Errorf(
				"Unable to read image file propery: %s: %w",
				file.Name(),
				e,
			)
		}

		reader, e := os.Open(filepath.Join(sourceDir, inputDir, file.Name()))
		if e != nil {
			return fmt.Errorf(
				"failed to open image file %s: %w",
				file.Name(),
				e,
			)
		}

		defer reader.Close()

		asset, err := immich.PostAsset(
			server,
			inputDir,
			file.Name(),
			reader,
			info.ModTime(),
		)
		if err != nil {
			slog.Error("failed to upload asset",
				slog.String("album", inputDir),
				slog.String("entry", file.Name()),
				slog.String("error", err.Error()),
			)
			continue
		}

		slog.Info("uploaded asset", slog.Any("asset", asset))
	}

	return nil
}
