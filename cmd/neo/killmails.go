package main

import (
	"strconv"
	"strings"

	"github.com/eveisesi/neo"
	core "github.com/eveisesi/neo/app"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func killmailCommands() []cli.Command {
	return []cli.Command{
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
			Name:  "backup",
			Usage: "Monitors a redis sorted set. As fully processed killmails populate the queue, backup pulls them off and pushes them to a digital ocean space",
			Action: func(c *cli.Context) error {
				app := core.New("killmail-backup", false)
				if !app.Config.SpacesEnabled {
					return cli.NewExitError("spaces is disabled. Exiting", 0)
				}
				app.Backup.Run(c.Int64("gLimit"), c.Int64("gSleep"))
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
		// cli.Command{
		// 	Name:  "stats",
		// 	Usage: "Monitors the stats queue and calculates stats as new killmails get processed",
		// 	Action: func(c *cli.Context) error {
		// 		app := core.New("killmail-stats", false)
		// 		_ = app.Stats.Run()
		// 		return nil
		// 	},
		// 	Subcommands: []cli.Command{
		// 		cli.Command{
		// 			Name:  "recalculate",
		// 			Usage: "Something Something",
		// 			Action: func(c *cli.Context) error {
		// 				app := core.New("killmail-recalculate", true)

		// 				var id int64
		// 				var entity string

		// 				dateStr := c.String("date")
		// 				now := time.Now()
		// 				date, err := time.Parse("20060102", dateStr)
		// 				if err != nil {
		// 					date = time.Date(now.Year(), now.Month(), now.Day()-90, 0, 0, 0, 0, time.UTC)
		// 				}

		// 				if c.Int64("id") != 0 {
		// 					id = c.Int64("id")
		// 				}
		// 				if c.String("entity") != "" {
		// 					for _, v := range app.Config.AllowedStatsEntities {
		// 						if v == c.String("entity") {
		// 							entity = c.String("entity")
		// 							break
		// 						}
		// 					}
		// 					if entity == "" {
		// 						app.Logger.Info("invalid type submitted. defaulting to empty string")
		// 					}
		// 				}

		// 				app.Stats.Recalculate(context.Background(), id, entity, date)

		// 				return nil

		// 			},
		// 			Flags: []cli.Flag{
		// 				cli.StringFlag{
		// 					Name:  "entity",
		// 					Usage: "declare the entity for this stats operations",
		// 				},
		// 				cli.Int64Flag{
		// 					Name:  "id",
		// 					Usage: "id of the type this stats operations is for",
		// 				},
		// 				cli.StringFlag{
		// 					Name:  "date",
		// 					Usage: "limit to a specific date. If ommitted, operation will be data from within the past 90 days (Format: YYYYMMDD)",
		// 				},
		// 			},
		// 		},
		// 	},
		// },
		cli.Command{
			Name:  "add",
			Usage: "Adds a Killmail ID and Hash to the queue",
			Action: func(c *cli.Context) error {

				in := c.String("in")
				delete := c.Bool("delete")

				app := core.New("killmail-add", false)
				entry := app.Logger.WithFields(logrus.Fields{
					"in": in,
				})

				inSlc := strings.Split(in, ":")
				id, err := strconv.ParseUint(inSlc[0], 10, 64)
				if err != nil {
					entry.WithError(err).Error("failed to parse id")
				}

				hash := inSlc[1]

				if delete {
					_, err := app.DB.Exec(`DELETE FROM killmails where id = ? AND hash = ?`, id, hash)
					if err != nil {
						entry.WithError(err).Fatal("failed to delete killmail with id and hash provided")
					}
				}

				app.Killmail.DispatchPayload(&neo.Message{ID: id, Hash: hash})

				return nil

			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "in",
					Usage:    "id:hash",
					Required: true,
				},
				cli.BoolFlag{
					Name:  "delete",
					Usage: "delete this killmail before dispatching",
				},
			},
		},
	}
}
