package backfill

import (
	"context"

	"github.com/urfave/cli/v3"
)

func Command(profile *string, displayLogLevel *string, fileLogLevel *string) *cli.Command {
	sourceDir := ""
	inputDir := ""
	dryRun := false
	archivePath := ""

	return &cli.Command{
		Name:  "backfill",
		Usage: "Reimport assets to backfill missing metadata.",
		Commands: []*cli.Command{
			{
				Name:  "directory",
				Usage: "Reimport assets in a directory.",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return backfillDirectory(
						ctx,
						*profile,
						*displayLogLevel,
						*fileLogLevel,
						sourceDir,
						inputDir,
						dryRun,
					)
				},
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
			}, {
				Name:  "archive",
				Usage: "Reimport assets in an archive file.",
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
					return backfillArchive(
						ctx,
						*profile,
						*displayLogLevel,
						*fileLogLevel,
						sourceDir,
						archivePath,
						dryRun,
					)
				},
			},
		},
	}
}
