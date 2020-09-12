package main

import (
	core "github.com/eveisesi/neo/app"
	"github.com/urfave/cli"
)

func test() cli.Command {
	return cli.Command{
		Name: "test",
		Action: func(c *cli.Context) error {
			app := core.New("f", false)

			startDate := c.String("startDate")
			endDate := c.String("endDate")
			incrementer := c.Int64("incrementer")
			stats := c.Bool("stats")

			app.History.Run(startDate, endDate, incrementer, stats)

			return nil
		},
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:     "startDate",
				Usage:    "Date to start the loop at when calling the zkillboard history api. (Format: YYYYMMDD)",
				Required: true,
			},
			cli.StringFlag{
				Name:     "endDate",
				Usage:    "Date to stop the history loop at when calling zkillboard history api. (Format: YYYYMMDD)",
				Required: true,
			},
			cli.Int64Flag{
				Name:     "incrementer",
				Usage:    "Direction to traverse dates in. Option are min and max. If min, script will start at the provided mindate and increment one day to max. If max, script will start at maxdate and decrement to min.",
				Required: true,
			},
			cli.BoolFlag{
				Name:  "stats",
				Usage: "Fetch Totals from ZKillboard and compare count to current db, but do not run fetch",
			},
		},
	}
}
