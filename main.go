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

	cmd := &cli.Command{
		Usage: "assets importer for immich.",
		Commands: []*cli.Command{
			{
				Name:  "archive",
				Usage: "Import assets from archives file in a directory.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "display-log-level",
						Value: "warn",
						Usage: "minimum log level on display",
					},
					&cli.StringFlag{
						Name:  "file-log-level",
						Value: "info",
						Usage: "minimum log level in log file.",
					},
					&cli.StringFlag{
						Name:  "profile",
						Value: "default",
						Usage: "profile of immich server.",
					},
					&cli.StringFlag{
						Name:    "source",
						Aliases: []string{"src"},
						Usage:   "source directory.",
					},
					&cli.BoolFlag{
						Name:  "force",
						Usage: "force processing the archive file even if an album with the same name exists.",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					err = SetupLog(
						cmd.Flags[0].Get().(string),
						cmd.Flags[1].Get().(string),
						os.Stdout, logFile,
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

					return archive.Process(
						url,
						c.ImmichAPIKey,
						cmd.Flags[3].Get().(string),
						cmd.Flags[4].Get().(bool),
					)
				},
			},
			{
				Name:  "directory",
				Usage: "Import assets from subdirectories of a directory.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "display-log-level",
						Value: "warn",
						Usage: "minimum log level on display",
					},
					&cli.StringFlag{
						Name:  "file-log-level",
						Value: "info",
						Usage: "minimum log level in log file.",
					},
					&cli.StringFlag{
						Name:  "profile",
						Value: "default",
						Usage: "profile of immich server.",
					},
					&cli.StringFlag{
						Name:    "source",
						Aliases: []string{"src"},
						Usage:   "source directory.",
					},
					&cli.BoolFlag{
						Name:  "force",
						Usage: "force processing the archive file even if an album with the same name exists.",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					err = SetupLog(
						cmd.Flags[0].Get().(string),
						cmd.Flags[1].Get().(string),
						os.Stdout, logFile,
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

					return directory.Process(
						url,
						c.ImmichAPIKey,
						cmd.Flags[3].Get().(string),
						cmd.Flags[4].Get().(bool),
					)
				},
			},
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
