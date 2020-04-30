package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/eveisesi/neo"
	"github.com/sirupsen/logrus"

	"github.com/inancgumus/screen"

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

					screen.Clear()
					screen.MoveTopLeft()

					esiPastFiveMinutes, err := app.Redis.ZCount(neo.REDIS_ESI_TRACKING_SUCCESS, strconv.FormatInt(time.Now().Add(time.Minute*-5).UnixNano(), 10), strconv.FormatInt(time.Now().UnixNano(), 10)).Result()
					if err != nil {
						return cli.NewExitError(err, 1)
					}

					fmt.Printf("%d: Successful ESI Calls in Past Five Minutes (%d)\t\t%d: Failed ESI Calls in the Past Five Minutes (%d)\n", esiPastFiveMinutes, esiPastFiveMinutes-prevEsiPastFiveMinutes, 0, 0)
					time.Sleep(time.Second * 2)
					prevEsiPastFiveMinutes = esiPastFiveMinutes

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

				for {

					status, err := app.Redis.Get(neo.REDIS_ESI_TRACKING_STATUS).Int64()
					if err != nil && err.Error() != neo.ErrRedisNil.Error() {
						app.Logger.WithError(err).Fatal("unexpected error encountered attempting to get tracking status from redis")
					}

					count, err := app.Redis.Get(neo.REDIS_ESI_ERROR_COUNT).Int64()
					if err != nil && err.Error() != neo.ErrRedisNil.Error() {
						app.Logger.WithError(err).Fatal("unexpected error encountered attempting to get error count from redis")
					}

					players, err := app.Redis.Get(neo.TQ_PLAYER_COUNT).Int64()
					if err != nil && err.Error() != neo.ErrRedisNil.Error() {
						app.Logger.WithError(err).Fatal("unexpected error encountered attempting to get error count from redis")
					}

					vip, err := app.Redis.Get(neo.TQ_VIP_MODE).Int64()
					if err != nil && err.Error() != neo.ErrRedisNil.Error() {
						app.Logger.WithError(err).Fatal("unexpected error encountered attempting to get error count from redis")
					}

					app.Logger.WithFields(logrus.Fields{
						neo.REDIS_ESI_TRACKING_STATUS: status,
						neo.REDIS_ESI_ERROR_COUNT:     count,
						neo.TQ_PLAYER_COUNT:           players,
						neo.TQ_VIP_MODE:               vip,
					}).Println()

					// Status:
					// Downtime: 3
					// Red: 2
					// Yellow: 1
					// Green: 0

					now := time.Now().In(time.UTC)

					if players < 100 {
						if status != neo.COUNT_STATUS_DOWNTIME {
							app.Logger.WithFields(logrus.Fields{
								neo.REDIS_ESI_TRACKING_STATUS: neo.COUNT_STATUS_DOWNTIME,
							}).Error("updating status in redis")
							app.Redis.Set(neo.REDIS_ESI_TRACKING_STATUS, neo.COUNT_STATUS_DOWNTIME, 0)
						}
					} else if vip > 0 {
						if status != neo.COUNT_STATUS_DOWNTIME {
							app.Logger.WithFields(logrus.Fields{
								neo.REDIS_ESI_TRACKING_STATUS: neo.COUNT_STATUS_DOWNTIME,
							}).Error("updating status in redis")
							app.Redis.Set(neo.REDIS_ESI_TRACKING_STATUS, neo.COUNT_STATUS_DOWNTIME, 0)
						}
					} else {
						if status == neo.COUNT_STATUS_DOWNTIME {
							if now.Unix() < start.Unix() || now.Unix() > end.Unix() {
								app.Logger.WithFields(logrus.Fields{
									neo.REDIS_ESI_TRACKING_STATUS: neo.COUNT_STATUS_GREEN,
								}).Info("updating status in redis")
								app.Redis.Set(neo.REDIS_ESI_TRACKING_STATUS, neo.COUNT_STATUS_GREEN, 0)
							}
						} else if status != neo.COUNT_STATUS_DOWNTIME {
							if now.Unix() >= start.Unix() && now.Unix() <= end.Unix() {
								app.Logger.WithFields(logrus.Fields{
									neo.REDIS_ESI_TRACKING_STATUS: neo.COUNT_STATUS_DOWNTIME,
								}).Info("updating status in redis")
								app.Redis.Set(neo.REDIS_ESI_TRACKING_STATUS, neo.COUNT_STATUS_DOWNTIME, 0)
							}
						} else if status == neo.COUNT_STATUS_RED {
							if count > 20 {
								app.Logger.WithFields(logrus.Fields{
									neo.REDIS_ESI_TRACKING_STATUS: neo.COUNT_STATUS_GREEN,
								}).Error("updating status in redis")
								app.Redis.Set(neo.REDIS_ESI_TRACKING_STATUS, neo.COUNT_STATUS_GREEN, 0)
							} else if count >= 10 && count <= 20 {
								app.Logger.WithFields(logrus.Fields{
									neo.REDIS_ESI_TRACKING_STATUS: neo.COUNT_STATUS_YELLOW,
								}).Warning("updating status in redis")
								app.Redis.Set(neo.REDIS_ESI_TRACKING_STATUS, neo.COUNT_STATUS_YELLOW, 0)
							}
						} else if status == neo.COUNT_STATUS_YELLOW {
							if count < 10 {
								app.Logger.WithFields(logrus.Fields{
									neo.REDIS_ESI_TRACKING_STATUS: neo.COUNT_STATUS_RED,
								}).Warning("updating status in redis")
								app.Redis.Set(neo.REDIS_ESI_TRACKING_STATUS, neo.COUNT_STATUS_RED, 0)
							} else if count >= 20 {
								app.Logger.WithFields(logrus.Fields{
									neo.REDIS_ESI_TRACKING_STATUS: neo.COUNT_STATUS_GREEN,
								}).Info("updating status in redis")
								app.Redis.Set(neo.REDIS_ESI_TRACKING_STATUS, neo.COUNT_STATUS_GREEN, 0)
							}
						} else if status == neo.COUNT_STATUS_GREEN {
							if count <= 20 {
								app.Logger.WithFields(logrus.Fields{
									neo.REDIS_ESI_TRACKING_STATUS: neo.COUNT_STATUS_YELLOW,
								}).Warning("updating status in redis")
								app.Redis.Set(neo.REDIS_ESI_TRACKING_STATUS, neo.COUNT_STATUS_YELLOW, 0)
							} else if count <= 10 {
								app.Logger.WithFields(logrus.Fields{
									neo.REDIS_ESI_TRACKING_STATUS: neo.COUNT_STATUS_RED,
								}).Warning("updating status in redis")
								app.Redis.Set(neo.REDIS_ESI_TRACKING_STATUS, neo.COUNT_STATUS_RED, 0)
							}
						}

					}

					if status > neo.COUNT_STATUS_GREEN && status < neo.COUNT_STATUS_DOWNTIME {
						ts, err := app.Redis.Get(neo.REDIS_ESI_ERROR_RESET).Int64()
						if err != nil {
							continue
						}

						if now.Unix() > ts && status != neo.COUNT_STATUS_GREEN {
							app.Logger.Info("set tracking green. error count has been reset")
							app.Redis.Set(neo.REDIS_ESI_TRACKING_STATUS, neo.COUNT_STATUS_GREEN, 0)
						}
					}
					time.Sleep(time.Second)
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
