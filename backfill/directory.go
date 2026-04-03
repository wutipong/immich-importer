package backfill

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/wutipong/immich-importer/archive"
	"github.com/wutipong/immich-importer/config"
	"github.com/wutipong/immich-importer/directory"
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

	if ctx.Err() != nil {
		err = fmt.Errorf("context error: %w", ctx.Err())
		return err
	}

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

	slog.Debug("processing path",
		slog.String("sourceDir", sourceDir),
		slog.String("inputDir", inputDir),
	)
	assetIds, err := directory.Process(ctx, server, sourceDir, inputDir)
	if err != nil {
		slog.Error(
			"failed upload assets.",
			slog.String("error", err.Error()),
			slog.String("sourceDir", sourceDir),
			slog.String("inputDir", inputDir),
		)
		if errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	}

	if len(assetIds) > 0 {
		slog.Info("creating album", slog.String("name", inputDir))
		createdAlbum, err := immich.CreateAlbum(
			ctx, server, inputDir, assetIds,
		)
		if err != nil {
			slog.Error("failed to create album", slog.String("error", err.Error()))
			return nil
		}

		slog.Info("created album", slog.Any("album", createdAlbum))
	} else {
		slog.Warn("directory does not contain any media files. Walking sub-directory next.")
	}

	err = filepath.WalkDir(filepath.Join(sourceDir, inputDir), func(path string, d os.DirEntry, err error,
	) error {
		if err != nil {
			slog.Warn(
				"failed to access path. skipping.",
				slog.String("path", path),
				slog.String("error", err.Error()),
			)
			return nil
		}

		slog.Debug("processing path",
			slog.String("path", path),
			slog.String("sourceDir", sourceDir),
			slog.String("inputDir", inputDir),
		)

		var assetIds []string

		albumPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			slog.Error(
				"failed to determine album name.",
				slog.String("error", err.Error()),
			)
			return nil
		}

		slog.Debug("processing path",
			slog.String("path", path),
			slog.String("albumPath", albumPath),
		)

		if d.IsDir() {
			assetIds, err = directory.Process(ctx, server, sourceDir, albumPath)
		} else {
			assetIds, err = archive.Process(ctx, server, sourceDir, albumPath)
		}

		if err != nil {
			slog.Error(
				"failed upload assets.",
				slog.String("error", err.Error()),
				slog.String("sourceDir", sourceDir),
				slog.String("albumPath", albumPath),
			)
			if errors.Is(err, context.Canceled) {
				return err
			}
			return nil
		}

		if len(assetIds) == 0 {
			slog.Debug(
				"no assets uploaded. skip create album.",
				slog.String("name", albumPath),
			)
			return nil
		}

		slog.Info("creating album", slog.String("name", albumPath))
		createdAlbum, err := immich.CreateAlbum(
			ctx, server, albumPath, assetIds,
		)
		if err != nil {
			slog.Error("failed to create album", slog.String("error", err.Error()))
			return nil
		}

		slog.Info("created album", slog.Any("album", createdAlbum))

		return nil
	})

	return nil
}
