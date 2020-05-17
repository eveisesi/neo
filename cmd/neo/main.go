package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/eveisesi/neo"
	"github.com/jedib0t/go-pretty/table"
	"github.com/pkg/errors"

	"github.com/inancgumus/screen"

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
					cron.WithSeconds(),
				)
				_, _ = c.AddFunc("0 10 11 * * *", func() {
					app := core.New()

					app.Market.FetchHistory(0)

				})

				_, _ = c.AddFunc("*/30 * * * * *", func() {

					app := core.New()
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
					app.Redis.Close()
					app.DB.Close()
				})

				_, _ = c.AddFunc("0 * * * * *", func() {
					app := core.New()

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
				_ = core.New().Killmail.Websocket()

				return nil
			},
		},
		cli.Command{
			Name: "monitor",
			Action: func(c *cli.Context) error {
				var err error
				app := core.New()

				var params = struct {
					SuccessfulESI     int64
					PrevSuccessfulESI int64

					FailedESI     int64
					PrevFailedESI int64

					ProcessingQueue     int64
					PrevProcessingQueue int64
				}{}
				for {

					screen.Clear()
					screen.MoveTopLeft()

					tw := table.NewWriter()
					params.SuccessfulESI, err = app.Redis.ZCount(neo.REDIS_ESI_TRACKING_SUCCESS, strconv.FormatInt(time.Now().Add(time.Minute*-5).UnixNano(), 10), strconv.FormatInt(time.Now().UnixNano(), 10)).Result()
					if err != nil {
						return cli.NewExitError(errors.Wrap(err, "failed to fetch successful esi calls"), 1)
					}

					params.FailedESI, err = app.Redis.ZCount(neo.REDIS_ESI_TRACKING_FAILED, strconv.FormatInt(time.Now().Add(time.Minute*-5).UnixNano(), 10), strconv.FormatInt(time.Now().UnixNano(), 10)).Result()
					if err != nil {
						return cli.NewExitError(errors.Wrap(err, "failed to fetch failed esi calls"), 1)
					}

					params.ProcessingQueue, err = app.Redis.ZCount(neo.QUEUES_KILLMAIL_PROCESSING, "-inf", "+inf").Result()
					if err != nil {
						return cli.NewExitError(errors.Wrap(err, "failed to fetch failed esi calls"), 1)
					}

					tw.AppendRows(
						[]table.Row{
							table.Row{
								fmt.Sprintf(
									"%d: Queue Processing (%d)",
									params.ProcessingQueue,
									params.ProcessingQueue-params.PrevProcessingQueue,
								),
								fmt.Sprintf(
									"%d: Successful ESI Call in Last Five Minutes (%d)",
									params.SuccessfulESI,
									params.SuccessfulESI-params.PrevSuccessfulESI,
								),
							},
							table.Row{
								"",
								fmt.Sprintf(
									"%d: Failed ESI Call in Last Five Minutes (%d)",
									params.FailedESI,
									params.FailedESI-params.PrevFailedESI,
								),
							},
						},
					)

					fmt.Println(tw.Render())

					time.Sleep(time.Second * 2)

					params.PrevSuccessfulESI = params.SuccessfulESI
					params.PrevFailedESI = params.FailedESI
					params.PrevProcessingQueue = params.ProcessingQueue
				}
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
			Name: "buildAutoCompleter",
			Action: func(c *cli.Context) error {

				app := core.New()

				err := app.Search.Build()
				if err != nil {
					return cli.NewExitError(err, 1)
				}

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
