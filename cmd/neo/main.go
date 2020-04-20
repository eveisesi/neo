package main

import (
	"log"
	"os"
	"time"

	core "github.com/eveisesi/neo/app"
	"github.com/eveisesi/neo/killmail/websocket"
	"github.com/eveisesi/neo/server"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/urfave/cli"
	"github.com/volatiletech/null"
)

var (
	app *cli.App
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("godotenv: ", err)
	}

	app = cli.NewApp()
	app.Name = "Neo Core"
	app.Usage = "Service that manages all services related to Neo and its stable operation"
	app.Version = "v0.0.1"
	app.Commands = []cli.Command{
		cli.Command{
			Name:  "import",
			Usage: "Listen to a Redis PubSub channel for killmail hashes. On Message receive, reach out to CCP for Killmail Data and process.",
			Action: func(c *cli.Context) error {
				app := core.New()
				channel := c.String("channel")
				limit := c.Int64("gLimit")
				sleep := c.Int64("gSleep")

				err := app.Killmail.Importer(channel, limit, sleep)
				if err != nil {
					return cli.NewExitError(err, 1)
				}

				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "channel",
					Usage:    "channel is the key to use when push killmail ids and hashes to redis",
					Required: true,
				},
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
			Name:  "history",
			Usage: "Reaches out to the Zkillboard API and downloads historical killmail hashes, then reaches out to CCP for Killmail Data",
			Action: func(c *cli.Context) error {
				app := core.New()
				channel := c.String("channel")
				date := null.NewString(c.String("date"), c.String("date") != "")

				err := app.Killmail.HistoryExporter(channel, date)
				if err != nil {
					return cli.NewExitError(err, 1)
				}

				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "channel",
					Usage:    "channel is the key to use when  pulling killmail ids and hashes from redis to be resolved and inserted into the database",
					Required: true,
				},
				cli.StringFlag{
					Name:  "date",
					Usage: "Date to use when request killmail hashes from zkillboard. (Format: YYYYMMDD)",
					// Required: true,
				},
			},
		},
		// cli.Command{
		// 	Name: "burner",
		// 	Action: func(c *cli.Context) error {
		// 		a := []int{1, 2, 3, 4, 5}
		// 		i := 2
		// 		// a = append(a[:1], a[2:]...)
		// 		spew.Dump(a)

		// 		copy(a[i:], a[i+1:])
		// 		a = a[:len(a)-1]
		// 		spew.Dump(a)

		// 		return nil
		// 	},
		// },
		cli.Command{
			Name:   "serve",
			Usage:  "Starts an HTTP Server to serve killmail data",
			Action: server.Action,
		},
		cli.Command{
			Name:  "market",
			Usage: "Opens a WSS Connection to ZKillboard and lsitens to the stream",
			Action: func(ctx *cli.Context) error {
				app := core.New()

				app.Market.FetchOrders()
				c := cron.New(
					cron.WithLocation(time.UTC),
					cron.WithLogger(
						cron.VerbosePrintfLogger(
							log.New(
								os.Stdout,
								"cron: ", log.LstdFlags,
							),
						),
					),
				)
				c.AddFunc("*/10 * * * *", func() {
					app.Market.FetchOrders()
				})

				c.Run()

				return nil
			},
		},
		cli.Command{
			Name:   "listen",
			Usage:  "Opens a WSS Connection to ZKillboard and lsitens to the stream",
			Action: websocket.Action,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "channel",
					Usage:    "channel is the key to use when pushing killmail ids and hashes to redis to be resolved and inserted into the database",
					Required: true,
				},
			},
		},
	}
}

func main() {
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
