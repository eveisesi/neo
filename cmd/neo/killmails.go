package main

import (
	"strconv"
	"strings"

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
				app := core.New(false)
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

				core.New(false).Backup.Run(c.Int64("gLimit"), c.Int64("gSleep"))

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
				delete := c.Bool("delete")

				app := core.New(false)
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

				app.Killmail.DispatchPayload(id, hash)

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
