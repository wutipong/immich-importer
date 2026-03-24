package config

import (
	"context"

	"github.com/urfave/cli/v3"
)

func Command(profile *string) *cli.Command {
	return &cli.Command{
		Name: "setup",
		Usage: "Setup configuration file interactively. " +
			"Existing configuration file will be overwritten.",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return SetupConfig(*profile)
		},
	}
}
