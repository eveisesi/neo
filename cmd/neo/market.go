package main

import (
	core "github.com/eveisesi/neo/app"
	"github.com/urfave/cli"
)

func marketCommands() []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "all",
			Usage: "Fetches Prices and History",
			Action: func(c *cli.Context) error {
				app := core.New(false)

				app.Market.FetchPrices()
				app.Market.FetchHistory()

				return nil
			},
		},
		cli.Command{
			Name:  "prices",
			Usage: "Fetches Prices",
			Action: func(c *cli.Context) error {
				app := core.New(false)

				app.Market.FetchPrices()

				return nil
			},
		},
		cli.Command{
			Name:  "history",
			Usage: "Fetches Market History",
			Action: func(c *cli.Context) error {
				app := core.New(false)

				app.Market.FetchHistory()

				return nil
			},
		},
	}
}
