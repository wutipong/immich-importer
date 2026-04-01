package config

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
	"github.com/wutipong/immich-importer/logging"
)

func Command(profile *string, displayLogLevel *string, fileLogLevel *string) *cli.Command {
	return &cli.Command{
		Name: "setup",
		Usage: "Setup configuration file interactively. " +
			"Existing configuration file will be overwritten.",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			err := logging.Setup(*profile, *displayLogLevel, true, *fileLogLevel)
			if err != nil {
				return fmt.Errorf("unable to setup log: %w", err)
			}
			defer logging.CleanUp()

			return SetupConfig(*profile)
		},
	}
}
