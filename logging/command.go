package logging

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"
)

func Command(profile *string, displayLogLevel *string, fileLogLevel *string) *cli.Command {
	keepLatestLogFiles := 0

	return &cli.Command{
		Name:  "log",
		Usage: "Log file-related commands.",
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

					entries, err := GetLogFileList(*profile)
					if err != nil {
						return fmt.Errorf("unable to get log files: %w", err)
					}

					if len(entries) == 0 {
						return fmt.Errorf("no log files found")
					}
					logDir, err := CreateLogDirectoryPath()
					if err != nil {
						return fmt.Errorf("unable to get log directory: %w", err)
					}

					println(filepath.Join(logDir, entries[0].Name()))
					return nil
				},
			}, {
				Name: "purge",
				Usage: "Permanently delete log files. " +
					"Use with caution, as this action cannot be undone.",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					err := Setup(*profile, *displayLogLevel, false, *fileLogLevel)
					if err != nil {
						return fmt.Errorf("unable to setup log: %w", err)
					}
					defer CleanUp()

					entries, err := GetLogFileList(*profile)
					if err != nil {
						return fmt.Errorf("unable to get log files: %w", err)
					}

					logDir, err := CreateLogDirectoryPath()
					if err != nil {
						return fmt.Errorf("unable to get log directory: %w", err)
					}

					for i, entry := range entries {
						if i < keepLatestLogFiles {
							continue
						}
						path := filepath.Join(logDir, entry.Name())
						err := os.Remove(path)
						if err != nil {
							return fmt.Errorf("unable to delete log file %s: %w", path, err)
						}
					}
					return nil
				},
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:        "keep-latest",
						Aliases:     []string{"keep"},
						Value:       0,
						Usage:       "Number of latest log files to keep. Older files will be deleted.",
						Destination: &keepLatestLogFiles,
					},
				},
			},
		},
	}
}
