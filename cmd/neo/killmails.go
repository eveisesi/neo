package main

import (
	"strconv"
	"strings"

	"github.com/eveisesi/neo"
	core "github.com/eveisesi/neo/app"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func killmail() cli.Command {
	return cli.Command{
		Name:  "killmail",
		Usage: "Parent command for all administrative task around killmails",
		Subcommands: []cli.Command{
			cli.Command{
				Name:  "import",
				Usage: "Listen to a Redis PubSub channel for killmail hashes. On Message receive, reach out to CCP for Killmail Data and process.",
				Action: func(c *cli.Context) error {
					app := core.New("killmail-import", false)
					limit := c.Int64("gLimit")
					sleep := c.Int64("gSleep")

					err := app.Killmail.Importer(limit, sleep)
					if err != nil {
						return cli.NewExitError(err, 1)
					}

					return nil
				},
				Flags: []cli.Flag{
					cli.Int64Flag{
						Name:     "gLimit",
						Usage:    "gLimit is the number of goroutines that the limiter should allow to be in flight at any one time",
						Required: true,
					},
					cli.Int64Flag{
						Name:     "gSleep",
						Usage:    "gSleep is the number of milliseconds the limiter will sleep between launching go routines when a slot is available",
						Required: true,
					},
				},
			},
			cli.Command{
				Name:  "add",
				Usage: "Adds a Killmail ID and Hash to the queue",
				Action: func(c *cli.Context) error {

					in := c.String("in")
					// delete := c.Bool("delete")

					app := core.New("killmail-add", false)
					entry := app.Logger.WithFields(logrus.Fields{
						"in": in,
					})

					inSlc := strings.Split(in, ":")
					id, err := strconv.ParseUint(inSlc[0], 10, 32)
					if err != nil {
						entry.WithError(err).Error("failed to parse id")
					}

					hash := inSlc[1]

					// if delete {
					// 	_, err := app.MySQLDB.Exec(`DELETE FROM killmails where id = ? AND hash = ?`, id, hash)
					// 	if err != nil {
					// 		entry.WithError(err).Fatal("failed to delete killmail with id and hash provided")
					// 	}
					// }

					app.Killmail.DispatchPayload(&neo.Message{ID: uint(id), Hash: hash})

					return nil

				},
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:     "in",
						Usage:    "id:hash",
						Required: true,
					},
					// cli.BoolFlag{
					// 	Name:  "delete",
					// 	Usage: "delete this killmail before dispatching",
					// },
				},
			},
			cli.Command{
				Name:  "history",
				Usage: "Reaches out to the Zkillboard API and downloads historical killmail hashes, then reaches out to CCP for Killmail Data",
				Action: func(c *cli.Context) error {
					app := core.New("killmail-history", false)
					maxdate := c.String("maxdate")
					mindate := c.String("mindate")
					threshold := c.Int64("threshold")
					direction := c.String("direction")
					datehold := c.Bool("datehold")
					overrideCurrent := c.Bool("overrideCurrent")

					err := app.Killmail.HistoryExporter(mindate, maxdate, direction, overrideCurrent, datehold, threshold)
					if err != nil {
						return cli.NewExitError(err, 1)
					}

					return cli.NewExitError(nil, 0)
				},
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "maxdate",
						Usage: "Date to start the loop at when calling the zkillboard history api. (Format: YYYYMMDD)",
					},
					cli.StringFlag{
						Name:     "mindate",
						Usage:    "Date to stop the history loop at when calling zkillboard history api. (Format: YYYYMMDD)",
						Required: true,
					},
					cli.StringFlag{
						Name:     "direction",
						Usage:    "Direction to traverse dates in. Option are min and max. If min, script will start at the provided mindate and increment one day to max. If max, script will start at maxdate and decrement to min.",
						Required: true,
					},
					cli.BoolFlag{
						Name:  "overrideCurrent",
						Usage: "Override cached current date so that the process can be reset if desired",
					},
					cli.BoolFlag{
						Name:  "datehold",
						Usage: "Hold after each date until the processing queue has reached a threshold. Threshold must be defined, else this command will be ignored",
					},
					cli.IntFlag{
						Name:  "threshold",
						Usage: "Threshold that the queue must be below process processing the next date",
					},
				},
			},
		},
	}

}
