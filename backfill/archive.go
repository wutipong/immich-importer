package backfill

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/wutipong/immich-importer/archive"
	"github.com/wutipong/immich-importer/config"
	"github.com/wutipong/immich-importer/immich"
	"github.com/wutipong/immich-importer/logging"
)

func backfillArchive(
	ctx context.Context, profile, displayLogLevel, fileLogLevel string,
	sourceDir, archivePath string,
	dryRun bool,
) error {
	err := logging.Setup(profile, displayLogLevel, true, fileLogLevel)
	if err != nil {
		return fmt.Errorf("unable to setup log: %w", err)
	}
	defer logging.CleanUp()

	if ctx.Err() != nil {
		return fmt.Errorf("context error: %w", ctx.Err())
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

	assetIds, err := archive.Process(ctx, server, sourceDir, archivePath)
	if err != nil {
		slog.Error(
			"failed upload assets.",
			slog.String("error", err.Error()),
			slog.String("sourceDir", sourceDir),
			slog.String("archivePath", archivePath),
		)
		if errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	}

	slog.Info("creating album", slog.String("name", archivePath))
	createdAlbum, err := immich.CreateAlbum(
		ctx,
		server, archivePath, assetIds,
	)
	if err != nil {
		slog.Error("failed to create album", slog.String("error", err.Error()))
		return nil
	}

	slog.Info("created album", slog.Any("album", createdAlbum))

	return nil
}
