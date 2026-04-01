package directory

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/wutipong/immich-importer/archive"
	"github.com/wutipong/immich-importer/config"
	"github.com/wutipong/immich-importer/immich"
	"github.com/wutipong/immich-importer/logging"
)

func Command(profile *string, displayLogLevel *string, fileLogLevel *string) *cli.Command {
	sourceDir := ""
	inputDir := ""
	dryRun := false

	return &cli.Command{
		Name:  "directory",
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
				Name:        "directory",
				Destination: &inputDir,
				Usage:       "directory file location, relative to source-dir. An album will be created from this directory. Also, the subdirectories will creates different albums.",
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

			DoCommand(server, sourceDir, inputDir)

			return nil
		},
	}
}

func DoCommand(server immich.ServerConfig, sourceDir string, inputDir string) error {
	slog.Debug("processing path",
		slog.String("sourceDir", sourceDir),
		slog.String("inputDir", inputDir),
	)
	assetIds, err := Process(server, sourceDir, inputDir)
	if err != nil {
		slog.Error(
			"failed upload assets.",
			slog.String("error", err.Error()),
		)
		return nil
	}

	if len(assetIds) > 0 {
		slog.Info("creating album", slog.String("name", inputDir))
		createdAlbum, err := immich.CreateAlbum(
			server, inputDir, assetIds,
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
			assetIds, err = Process(server, sourceDir, albumPath)
		} else {
			assetIds, err = archive.Process(server, sourceDir, albumPath)
		}

		if err != nil {
			slog.Error(
				"failed upload assets.",
				slog.String("error", err.Error()),
			)
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
			server, albumPath, assetIds,
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
