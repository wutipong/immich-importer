package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
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
			config.Command(&profile, &displayLogLevelStr, &fileLogLevelStr),
			run.Command(&profile, &displayLogLevelStr, &fileLogLevelStr),
			archive.Command(&profile, &displayLogLevelStr, &fileLogLevelStr),
			directory.Command(&profile, &displayLogLevelStr, &fileLogLevelStr),
			merge.Command(&profile, &displayLogLevelStr, &fileLogLevelStr),
			logging.Command(&profile, &displayLogLevelStr, &fileLogLevelStr),
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Print version information",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					version := "unknown"
					if info, ok := debug.ReadBuildInfo(); ok {
						version = info.Main.Version
					} else {
						fmt.Println("unknown")
					}
					fmt.Printf("immich-importer version: %s\n", version)
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("application ended with error", slog.String("error", err.Error()))
		return
	}
}
