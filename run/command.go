package run

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/wutipong/immich-importer/archive"
	"github.com/wutipong/immich-importer/config"
	"github.com/wutipong/immich-importer/directory"
	"github.com/wutipong/immich-importer/immich"
)

func Command(profile *string) *cli.Command {
	sourceDir := ""
	force := false
	dryRun := false
	disableDirectory := true
	disableArchive := true

	return &cli.Command{
		Name:  "run",
		Usage: "perform importing assets.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "source-dir",
				Aliases:     []string{"src", "source"},
				Usage:       "Source directory.",
				Destination: &sourceDir,
				Required:    true,
			},
			&cli.BoolFlag{
				Name:        "force",
				Value:       false,
				Usage:       "Force processing album even if an album with the same name exists.",
				Destination: &force,
			},
			&cli.BoolFlag{
				Name:        "dry-run",
				Value:       false,
				Usage:       "Processing assets without working with the Immich server.",
				Destination: &dryRun,
				Category:    "Processing",
			},
			&cli.BoolFlag{
				Name:        "disable-directory",
				Value:       false,
				Usage:       "Disable processing media files in directories.",
				Destination: &disableDirectory,
				Category:    "Processing",
			},
			&cli.BoolFlag{
				Name:        "disable-archive",
				Value:       false,
				Usage:       "Disable processing media files in archive files.",
				Destination: &disableArchive,
				Category:    "Processing",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
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

			return Process(
				server,
				sourceDir,
				force,
				!disableDirectory,
				!disableArchive,
			)
		},
	}
}

func Process(
	server immich.ServerConfig,
	sourceDir string,
	force bool,
	processDirectory bool,
	processArchive bool,

) error {
	var albums []immich.AlbumResponseDto
	var err error

	if !force {
		albums, err = immich.GetAlbums(server)
		if err != nil {
			err = fmt.Errorf("unable to retrieved existing albums: %w", err)
			return err
		}
		slog.Debug("albums", slog.Any("existing albums", albums))
	}
	err = filepath.WalkDir(sourceDir,
		func(path string, d os.DirEntry, err error,
		) error {
			if err != nil {
				slog.Warn(
					"failed to access path. skipping.",
					slog.String("path", path),
					slog.String("error", err.Error()),
				)
				return nil
			}

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

			matchingAlbums := slices.DeleteFunc(slices.Clone(albums),
				func(album immich.AlbumResponseDto) bool {
					return album.AlbumName != albumPath
				})

			slog.Debug("matching albums",
				slog.Any("album", albumPath),
				slog.Any("existing", albums),
				slog.Any("matchalbums", matchingAlbums),
			)

			if !force && len(matchingAlbums) > 0 {
				slog.Warn(
					"album already exists. skipping.",
					slog.String("name", albumPath),
				)
				return nil
			}

			if d.IsDir() {
				if !processDirectory {
					slog.Debug("skipping directory",
						slog.String("path", path),
					)
					return nil
				}
				assetIds, err = directory.Process(server, sourceDir, albumPath)
			} else {
				if !processArchive {
					slog.Debug("skipping file",
						slog.String("path", path),
					)
					return nil
				}

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

			if len(matchingAlbums) > 0 {
				slog.Info(
					"album already exists. update existing album.",
					slog.String("name", albumPath),
				)

				var albumIds []string
				for _, album := range matchingAlbums {
					albumIds = append(albumIds, album.ID)
				}

				err = immich.AddAssetsToAlbum(server, albumIds, assetIds)
				if err != nil {
					slog.Error(
						"failed to add assets to album",
						slog.String("error", err.Error()),
					)
				}
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

	return err
}
