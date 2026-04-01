package logging

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"
)

func Command(profile *string, displayLogLevel *string, fileLogLevel *string) *cli.Command {
	return &cli.Command{
		Name:  "log",
		Usage: "Logging related commands ",
		Commands: []*cli.Command{
			{
				Name:  "location",
				Usage: "Show log files location.",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					err := Setup(*profile, *displayLogLevel, false, *fileLogLevel)
					if err != nil {
						return fmt.Errorf("unable to setup log: %w", err)
					}
					defer CleanUp()

					path, err := CreateLogDirectoryPath()
					if err != nil {
						return err
					}

					println(path)
					return nil
				},
			}, {
				Name:  "latest",
				Usage: "Show the latest log file location.",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					err := Setup(*profile, *displayLogLevel, false, *fileLogLevel)
					if err != nil {
						return fmt.Errorf("unable to setup log: %w", err)
					}
					defer CleanUp()

					path, err := CreateLogDirectoryPath()
					if err != nil {
						return fmt.Errorf("unable to create log file path: %w", err)
					}

					entries, err := os.ReadDir(path)
					if err != nil {
						return fmt.Errorf("unable to read log directory: %w", err)
					}

					if len(entries) == 0 {
						return fmt.Errorf("no log files found")
					}

					latest := entries[0]
					latestInfo, err := latest.Info()
					if err != nil {
						return fmt.Errorf("unable to get log file info: %w", err)
					}
					for _, entry := range entries {
						if entry.IsDir() {
							continue
						}
						info, err := entry.Info()
						if err != nil {
							continue
						}
						if info.ModTime().After(latestInfo.ModTime()) {
							latest = entry
							latestInfo = info
						}
					}

					println(filepath.Join(path, latest.Name()))
					return nil
				},
			},
		},
	}
}
