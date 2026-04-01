package archive

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/wutipong/immich-importer/config"
	"github.com/wutipong/immich-importer/immich"
	"github.com/wutipong/immich-importer/logging"
)

func Command(profile *string, displayLogLevel *string, fileLogLevel *string) *cli.Command {
	sourceDir := ""
	archivePath := ""
	dryRun := false

	return &cli.Command{
		Name:  "archive",
		Usage: "create an album from an archive file.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "source-dir",
				Aliases:     []string{"src", "source"},
				Destination: &sourceDir,
				Usage:       "Directory that contains the archive.",
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "archive",
				Destination: &archivePath,
				Usage:       "Archive file location, relative to source-dir. Will be used as the album name.",
				Required:    true,
			},
			&cli.BoolFlag{
				Name:        "dry-run",
				Value:       false,
				Usage:       "Processing assets without working with the Immich server.",
				Destination: &dryRun,
				Category:    "Processing",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			err := logging.Setup(*profile, *displayLogLevel, true, *fileLogLevel)
			if err != nil {
				return fmt.Errorf("unable to setup log: %w", err)
			}
			defer logging.CleanUp()

			c, err := config.LoadConfig(*profile)
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

			assetIds, err := Process(server, sourceDir, archivePath)
			if err != nil {
				slog.Error(
					"failed upload assets.",
					slog.String("error", err.Error()),
				)
				return nil
			}

			slog.Info("creating album", slog.String("name", archivePath))
			createdAlbum, err := immich.CreateAlbum(
				server, archivePath, assetIds,
			)
			if err != nil {
				slog.Error("failed to create album", slog.String("error", err.Error()))
				return nil
			}

			slog.Info("created album", slog.Any("album", createdAlbum))

			return nil
		},
	}
}
