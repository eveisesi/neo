package main

import (
	"context"

	core "github.com/eveisesi/neo/app"
	"github.com/urfave/cli"
)

func alliances() cli.Command {
	return cli.Command{
		Name:  "alliances",
		Usage: "Actions to perform against a specific or a collection of alliances",
		Subcommands: []cli.Command{
			{
				Name:  "updateExpired",
				Usage: "Queries for expired records and attempts to update them",
				Action: func(c *cli.Context) error {
					app := core.New("f", false)

					app.Alliance.UpdateExpired(context.Background())

					return nil
				},
			},
		},
	}
}
