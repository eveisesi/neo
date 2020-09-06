package main

import (
	"context"

	core "github.com/eveisesi/neo/app"
	"github.com/urfave/cli"
)

func corporations() cli.Command {
	return cli.Command{
		Name:  "corporations",
		Usage: "Actions to perform against a specific or a collection of corporations",
		Subcommands: []cli.Command{
			{
				Name:  "updateExpired",
				Usage: "Queries for expired records and attempts to update them",
				Action: func(c *cli.Context) error {
					app := core.New("f", false)

					app.Corporation.UpdateExpired(context.Background())

					return nil
				},
			},
		},
	}
}
