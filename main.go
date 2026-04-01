package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/urfave/cli/v3"
	"github.com/wutipong/immich-importer/archive"
	"github.com/wutipong/immich-importer/config"
	"github.com/wutipong/immich-importer/directory"
	"github.com/wutipong/immich-importer/logging"
	"github.com/wutipong/immich-importer/merge"
	"github.com/wutipong/immich-importer/run"
)

func main() {
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      slog.LevelError,
		TimeFormat: time.Kitchen,
	})))

	displayLogLevelStr := "warn"
	fileLogLevelStr := "info"
	profile := "default"

	cmd := &cli.Command{
		Usage: "Import assets from subdirectories and archives file in a directory.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "display-log",
				Value:       "warn",
				Usage:       "Minimum log-level on display (debug, info, warn, error).",
				Destination: &displayLogLevelStr,
				Category:    "Logging",
			},
			&cli.StringFlag{
				Name:        "file-log",
				Value:       "info",
				Usage:       "Minimum log-level in log file (debug, info, warn, error).",
				Destination: &fileLogLevelStr,
				Category:    "Logging",
			},
			&cli.StringFlag{
				Name:        "profile",
				Value:       "default",
				Usage:       "profile of immich server.",
				Destination: &profile,
				Category:    "Immich Server",
			},
		},
		Commands: []*cli.Command{
			config.Command(&profile),
			run.Command(&profile),
			archive.Command(&profile),
			directory.Command(&profile),
			merge.Command(&profile),
			logging.Command(&profile),
		},
		Before: func(ctx context.Context, c *cli.Command) (ctx2 context.Context, err error) {

			err = logging.Setup(profile, displayLogLevelStr, true, fileLogLevelStr)
			if err != nil {
				slog.Error("unable to setup logging system", slog.String("error", err.Error()))
			}

			slog.Info("command started", slog.Any("command", c))
			return
		}, After: func(ctx context.Context, c *cli.Command) error {
			return logging.CleanUp()
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("application ended with error", slog.String("error", err.Error()))
		return
	}
}
