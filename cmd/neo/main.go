package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	core "github.com/eveisesi/neo/app"
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
		cli.Command{
			Name:   "serve",
			Usage:  "Starts an HTTP Server to serve killmail data",
			Action: server.Action,
		},
		cli.Command{
			Name:  "cron",
			Usage: "Spins up the crons",
			Action: func(ctx *cli.Context) error {

				if ctx.Bool("now") {
					from := 0
					if ctx.Int("from") > 0 {
						from = ctx.Int("from")
					}
					app := core.New()

					app.Market.FetchHistory(from)
				}

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
				_, _ = c.AddFunc("10 11 * * *", func() {
					app := core.New()

					app.Market.FetchHistory(0)

				})

				_, _ = c.AddFunc("* * * * *", func() {
					app := core.New()

					ts := time.Now().Add(time.Minute * -6).UnixNano()
					count, err := app.Redis.ZRemRangeByScore("esi:tracking:success", "-inf", strconv.FormatInt(ts, 10)).Result()
					if err != nil {
						app.Logger.WithError(err).Error("failed to fetch current count of esi success set from redis")
						return
					}

					app.Logger.WithField("removed", count).Info("successfully cleared keys from success queue")
					app.Redis.Close()
					app.DB.Close()
				})

				c.Run()

				return nil
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "now",
					Usage: "fetch orders immediately, then initiate the cron",
				},
				cli.IntFlag{
					Name:  "from",
					Usage: "Group ID to start fetch from",
				},
			},
		},
		cli.Command{
			Name:  "listen",
			Usage: "Opens a WSS Connection to ZKillboard and lsitens to the stream",
			Action: func(c *cli.Context) error {
				_ = core.New().Killmail.Websocket(c.String("channel"))

				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "channel",
					Usage:    "channel is the key to use when pushing killmail ids and hashes to redis to be resolved and inserted into the database",
					Required: true,
				},
			},
		},
		cli.Command{
			Name: "monitor",
			Action: func(c *cli.Context) error {

				app := core.New()
				prevEsiPastFiveMinutes := int64(0)
				for {
					esiPastFiveMinutes, err := app.Redis.ZCount("esi:tracking:success", strconv.FormatInt(time.Now().Add(time.Minute*-5).UnixNano(), 10), strconv.FormatInt(time.Now().UnixNano(), 10)).Result()
					if err != nil {
						return cli.NewExitError(err, 1)
					}

					fmt.Printf("%d: Successful ESI Call in Past Five Minutes (%d)\n", esiPastFiveMinutes, esiPastFiveMinutes-prevEsiPastFiveMinutes)
					time.Sleep(time.Second * 2)
					prevEsiPastFiveMinutes = esiPastFiveMinutes

				}
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
