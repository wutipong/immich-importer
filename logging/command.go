package logging

import (
	"context"
	"fmt"
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
			},
		},
	}
}
