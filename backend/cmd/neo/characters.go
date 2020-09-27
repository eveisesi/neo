package main

import (
	"context"

	core "github.com/eveisesi/neo/app"
	"github.com/urfave/cli"
)

func characters() cli.Command {
	return cli.Command{
		Name:  "characters",
		Usage: "Actions to perform against a specific or a collection of characters",
		Subcommands: []cli.Command{
			{
				Name:  "updateExpired",
				Usage: "Queries for expired records and attempts to update them",
				Action: func(c *cli.Context) error {
					app := core.New("f", false)

					app.Character.UpdateExpired(context.Background())

					return nil
				},
			},
		},
	}
}
