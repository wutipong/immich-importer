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
)

func Command(profile *string) *cli.Command {
	pattern := ""
	dryRun := false
	album := ""

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
				Usage:       "Regex pattern to match name of albums to merge.",
				Required:    true,
			},
			&cli.BoolFlag{
				Name:        "dry-run",
				Value:       false,
				Usage:       "Processing assets without crete an output.",
				Destination: &dryRun,
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

			err = Process(server, album, pattern)
			if err != nil {
				return fmt.Errorf("failed to merge albums: %w", err)
			}

			return nil
		},
	}
}
