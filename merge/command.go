package merge

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/wutipong/immich-importer/config"
	"github.com/wutipong/immich-importer/immich"
	"github.com/wutipong/immich-importer/logging"
)

func Command(profile *string, displayLogLevel *string, fileLogLevel *string) *cli.Command {
	pattern := ""
	dryRun := false
	album := ""
	disableDelete := false

	return &cli.Command{
		Name:  "merge",
		Usage: "create an album by merging multiple albums.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "album",
				Destination: &album,
				Usage:       "Album name to create.",
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "pattern",
				Destination: &pattern,
				Usage:       "Regex pattern to match name of albums to merge. Use Golang/RE2 pattern.",
				Required:    true,
			},
			&cli.BoolFlag{
				Name:        "disable-deletion",
				Destination: &disableDelete,
				Usage:       "Do not delete existing albums",
			},
			&cli.BoolFlag{
				Name:        "dry-run",
				Value:       false,
				Usage:       "Processing assets without working with the Immich server.",
				Destination: &dryRun,
				Category:    "Processing",
			},
		},
		UsageText: `Merge multiple albums that matches the given pattern into a new album. The pattern is in Google's RE2 syntax.
For example, to merge any albums with name begins with 'abc', use the pattern (without quote)'^abc.*$.
See https://pkg.go.dev/regexp/syntax for information on patterns.`,
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

			err = Process(ctx, server, album, pattern, !disableDelete)
			if err != nil {
				return fmt.Errorf("failed to merge albums: %w", err)
			}

			return nil
		},
	}
}
