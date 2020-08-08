package main

import (
	"context"

	core "github.com/eveisesi/neo/app"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/urfave/cli"
)

func marketCommands() []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "all",
			Usage: "Fetches Prices and History",
			Action: func(c *cli.Context) error {
				app := core.New("market", false)
				txn := app.NewRelic.StartTransaction(app.Label)
				defer txn.End()
				ctx := newrelic.NewContext(context.Background(), txn)
				app.Market.FetchPrices(ctx)
				app.Market.FetchHistory(ctx)

				return nil
			},
		},
		cli.Command{
			Name:  "prices",
			Usage: "Fetches Prices",
			Action: func(c *cli.Context) error {
				app := core.New("market-prices", false)
				txn := app.NewRelic.StartTransaction(app.Label)
				defer txn.End()
				ctx := newrelic.NewContext(context.Background(), txn)
				app.Market.FetchPrices(ctx)

				return nil
			},
		},
		cli.Command{
			Name:  "history",
			Usage: "Fetches Market History",
			Action: func(c *cli.Context) error {
				app := core.New("market-history", false)
				txn := app.NewRelic.StartTransaction(app.Label)
				defer txn.End()
				ctx := newrelic.NewContext(context.Background(), txn)
				app.Market.FetchHistory(ctx)

				return nil
			},
		},
	}
}
