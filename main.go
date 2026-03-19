package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"github.com/urfave/cli/v3"
	"github.com/wutipong/immich-importer/archive"
	"github.com/wutipong/immich-importer/config"
	"github.com/wutipong/immich-importer/directory"
	"github.com/wutipong/immich-importer/immich"
)

func main() {
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      slog.LevelError,
		TimeFormat: time.Kitchen,
	})))

	logFile, err := os.OpenFile(
		"immich-importer.log",
		os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
		0644,
	)
	if err != nil {
		slog.Error(
			"unable to open log file to write.",
			slog.String("error", err.Error()),
		)
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Error(
			"failed to get user home directory",
			slog.String("error", err.Error()),
		)
		return
	}

	configPath := filepath.Join(homeDir, ".immich-importer", "config.yaml")

	displayLogLevelStr := "warn"
	fileLogLevelStr := "info"
	profile := "default"
	sourceDir := ""
	force := false
	dryRun := false
	disableDirectory := true
	disableArchive := true

	cmd := &cli.Command{
		Usage: "Import assets from subdirectories and archives file in a directory.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "display-log",
				Value:       "warn",
				Usage:       "minimum log-level on display (debug, info, warn, error).",
				Destination: &displayLogLevelStr,
				Category:    "Logging",
			},
			&cli.StringFlag{
				Name:        "file-log",
				Value:       "info",
				Usage:       "minimum log-level in log file (debug, info, warn, error).",
				Destination: &fileLogLevelStr,
				Category:    "Logging",
			},
			&cli.StringFlag{
				Name:        "profile",
				Value:       "default",
				Usage:       "profile of immich server.",
				Destination: &profile,
			},
			&cli.StringFlag{
				Name:        "source",
				Aliases:     []string{"src"},
				Usage:       "source directory.",
				Destination: &sourceDir,
				Required:    true,
			},
			&cli.BoolFlag{
				Name:        "force",
				Value:       false,
				Usage:       "force processing album even if an album with the same name exists.",
				Destination: &force,
			},
			&cli.BoolFlag{
				Name:        "dry-run",
				Value:       false,
				Usage:       "processing directory without actually creating assets.",
				Destination: &dryRun,
				Category:    "Processing",
			},
			&cli.BoolFlag{
				Name:        "disable-directory",
				Value:       false,
				Usage:       "disable processing media files in directories.",
				Destination: &disableDirectory,
				Category:    "Processing",
			},
			&cli.BoolFlag{
				Name:        "disable-archive",
				Value:       false,
				Usage:       "disable processing media files in archive files.",
				Destination: &disableArchive,
				Category:    "Processing",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			err = SetupLog(
				displayLogLevelStr,
				fileLogLevelStr,
				os.Stdout,
				logFile,
			)
			if err != nil {
				return fmt.Errorf("unable to setup logging system. => %w", err)
			}

			c, err := config.LoadConfig(cmd.Flags[2].Get().(string), configPath)
			if err != nil {
				return fmt.Errorf("unable to load configuration. => %w", err)
			}

			slog.Info("Immich instance",
				slog.String("url", c.ImmichURL),
				slog.String("api_key",
					strings.Repeat("*", len(c.ImmichAPIKey)),
				),
			)

			url, err := url.Parse(c.ImmichURL)
			if err != nil {
				return fmt.Errorf("invalid immich url. => %w", err)
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

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("application ended with error", slog.String("error", err.Error()))
	}
}

func ParseLogLevel(levelStr string) (level slog.Level, err error) {
	switch strings.ToLower(strings.Trim(levelStr, "")) {
	case "debug":
		level = slog.LevelDebug
		return

	case "info":
		level = slog.LevelInfo
		return

	case "warn":
		level = slog.LevelWarn
		return

	case "error":
		level = slog.LevelError
		return
	}

	err = fmt.Errorf("invalid log levl: '%s'", levelStr)
	return
}

func SetupLog(
	dispLogLevelStr string,
	fileLogLevelStr string,
	dispWriter io.Writer,
	fileWriter io.Writer,
) error {
	fileLevel, err := ParseLogLevel(fileLogLevelStr)
	if err != nil {
		err = fmt.Errorf("unable to parse log level for log file. => %w", err)
		return err
	}

	dispLevel, err := ParseLogLevel(dispLogLevelStr)
	if err != nil {
		err = fmt.Errorf("unable to parse log level for log file. => %w", err)
		return err
	}

	tintHandler := tint.NewHandler(dispWriter, &tint.Options{
		Level:      dispLevel,
		TimeFormat: time.Kitchen,
	})
	jsonHandler := slog.NewJSONHandler(fileWriter, &slog.HandlerOptions{
		Level: fileLevel,
	})

	slog.SetDefault(slog.New(
		slog.NewMultiHandler(jsonHandler, tintHandler),
	))

	return nil
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
			err = fmt.Errorf("unable to retrieved existing albums. => %w", err)
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

			if d.IsDir() {
				if !processDirectory {
					slog.Debug("skipping directory",
						slog.String("path", path),
					)
					return nil
				}
				assetIds, err = directory.Process(server, sourceDir, albumPath)
			} else {
				if !archive.IsArchive(filepath.Ext(path)) {
					return nil
				}
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
