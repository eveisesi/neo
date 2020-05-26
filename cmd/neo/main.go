package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/eveisesi/neo"

	core "github.com/eveisesi/neo/app"
	"github.com/eveisesi/neo/server"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/urfave/cli"
)

var (
	app *cli.App
)

func init() {
	_ = godotenv.Load(".env")

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
			Name:  "history",
			Usage: "Reaches out to the Zkillboard API and downloads historical killmail hashes, then reaches out to CCP for Killmail Data",
			Action: func(c *cli.Context) error {
				app := core.New()
				maxdate := c.String("maxdate")
				mindate := c.String("mindate")

				err := app.Killmail.HistoryExporter(mindate, maxdate)
				if err != nil {
					return cli.NewExitError(err, 1)
				}

				return cli.NewExitError(nil, 0)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "maxdate",
					Usage:    "Date to start the loop at when calling the zkillboard history api. (Format: YYYYMMDD)",
					Required: true,
				},
				cli.StringFlag{
					Name:     "mindate",
					Usage:    "Date to stop the history loop at when calling zkillboard history api. (Format: YYYYMMDD)",
					Required: true,
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
					app := core.New()

					app.Market.FetchPrices()

					app.Market.FetchHistory()
				}
				app := core.New()

				c := cron.New(
					cron.WithLocation(time.UTC),
					cron.WithLogger(
						cron.PrintfLogger(
							// log.New(
							// 	os.Stdout,
							// 	"cron: ", log.LstdFlags,
							// ),
							app.Logger,
						),
					),
					cron.WithSeconds(),
				)

				_, _ = c.AddFunc("0 10 11 * * *", func() {

					app.Logger.Info("starting fetch prices")
					app.Market.FetchPrices()
					app.Logger.Info("done with fetch prices")
					app.Logger.Info("starting fetch history ")
					app.Market.FetchHistory()
					app.Logger.Info("done with fetch history ")

				})

				_, _ = c.AddFunc("*/30 * * * * *", func() {

					app.Logger.Info("checking tq server status")

					serverStatus, m := app.ESI.GetStatus()
					if m.IsError() {
						app.Logger.WithError(m.Msg).Error("Failed to fetch tq server status from ESI")
						return
					}

					if m.Code != 200 {
						app.Logger.WithField("code", m.Code).Error("unable to acquire tq server status")
						return
					}

					app.Redis.Set(neo.TQ_PLAYER_COUNT, serverStatus.Players, 0)
					app.Redis.Set(neo.TQ_VIP_MODE, serverStatus.VIP.Bool, 0)

					app.Logger.Info("done checking tq server status")

				})

				_, _ = c.AddFunc("0 * * * * *", func() {

					app.Logger.Info("starting esi tracking set janitor")

					ts := time.Now().Add(time.Minute * -6).UnixNano()
					count := int64(0)
					a, err := app.Redis.ZRemRangeByScore(neo.REDIS_ESI_TRACKING_SUCCESS, "-inf", strconv.FormatInt(ts, 10)).Result()
					if err != nil {
						app.Logger.WithError(err).Error("failed to fetch current count of esi success set from redis")
						return
					}
					count += a
					b, err := app.Redis.ZRemRangeByScore(neo.REDIS_ESI_TRACKING_FAILED, "-inf", strconv.FormatInt(ts, 10)).Result()
					if err != nil {
						app.Logger.WithError(err).Error("failed to fetch current count of esi success set from redis")
						return
					}
					count += b

					app.Logger.WithField("removed", count).Info("successfully cleared keys from success queue")
					app.Logger.Info("stopping esi tracking set janitor")

				})

				_, _ = c.AddFunc("0 0 11 * * *", func() {
					app := core.New()

					app.Logger.Info("rebuilding autocompleter index")

					err := app.Search.Build()
					if err != nil {
						app.Logger.WithError(err).Error("failed to rebuild autocompleter index")
					}

					app.Logger.Info("done rebuilding autocompleter index")
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
				_ = core.New().Killmail.Websocket()

				return nil
			},
		},
		cli.Command{
			Name: "top",
			Action: func(c *cli.Context) error {
				return core.New().Top.Run()
			},
		},
		cli.Command{
			Name: "tracking",
			Action: func(c *cli.Context) error {

				app := core.New()

				beginning := time.Now().In(time.UTC)
				start := time.Date(beginning.Year(), beginning.Month(), beginning.Day(), 10, 58, 0, 0, time.UTC)
				end := time.Date(beginning.Year(), beginning.Month(), beginning.Day(), 11, 25, 0, 0, time.UTC)

				app.Tracker.Run(start, end)

				return nil
			},
		},
		cli.Command{
			Name:        "autocompleter",
			Description: "Manually rebuild the autocompleter index",
			Action: func(c *cli.Context) error {

				app := core.New()

				err := app.Search.Build()
				if err != nil {
					return cli.NewExitError(err, 1)
				}

				return nil
			},
		},
		cli.Command{
			Name:        "notifications",
			Description: "Notifications subscribe to the a Redis PubSub. When the importer detects a killmail with a value greater than the configured notification value, it publishes the id and hash to this pubsub and this service will format the message for slack and post the killmail to slack",
			Action: func(c *cli.Context) error {
				app := core.New()

				app.Notification.Run()
				return nil
			},
		},
		cli.Command{
			Name:        "updater",
			Description: "Updater ensures that all updatable records in the database are update date according to their CacheUntil timestamp.",
			Action: func(c *cli.Context) error {
				app := core.New()
				var ctx = context.Background()

				ch := make(chan int, 3)

				go app.Character.UpdateExpired(ctx)
				go app.Corporation.UpdateExpired(ctx)
				go app.Alliance.UpdateExpired(ctx)

				<-ch

				return nil

			},
		},
		cli.Command{
			Name: "migrate",
			Action: func(c *cli.Context) error {

				app := core.New()

				app.Logger.Info("initialize migrations")

				err := app.Migration.Init()
				if err != nil {
					return cli.NewExitError(err, 1)
				}

				app.Logger.Info("migrations initialized")

				app.Logger.Info("running migrations")

				app.Migration.Run()

				app.Logger.Info("migrations run successfully. exiting application")
				time.Sleep(time.Second * 2)

				return nil
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
