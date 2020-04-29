package main

import (
	"log"
	"os"
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
		// cli.Command{
		// 	Name: "fill-the-gap",
		// 	Action: func(c *cli.Context) error {
		// 		app := core.New()

		// 		unique := make([]int, 0)
		// 		err = app.DB.Select(&unique, `
		// 			SELECT DISTINCT(type_id) FROM prices ORDER BY type_id ASC
		// 		`)

		// 		limit := limiter.NewConcurrencyLimiter(10)

		// 		start := time.Date(2018, 03, 05, 0, 0, 0, 0, time.UTC)
		// 		end := time.Date(2019, 03, 01, 0, 0, 0, 0, time.UTC)

		// 		for _, t := range unique {
		// 			if t < 20929 {
		// 				continue
		// 			}
		// 			limit.Execute(func() {
		// 				x := t
		// 				url := fmt.Sprintf("https://zkillboard.com/api/prices/%d/", x)
		// 				res, err := http.Get(url)
		// 				if err != nil {
		// 					app.Logger.WithError(err).WithField("type_id", x).Error("failed to make request for prices")
		// 					return
		// 				}

		// 				results := make(map[string]float64)
		// 				_ = json.NewDecoder(res.Body).Decode(&results)
		// 				delete(results, "currentPrice")
		// 				delete(results, "typeID")
		// 				args := make([]string, 0)
		// 				params := make([]interface{}, 0)
		// 				arg := "(?, ?, ?, NOW(), NOW())"
		// 				for i, v := range results {

		// 					date, err := time.ParseInLocation("2006-01-02", i, time.UTC)
		// 					if err != nil {
		// 						app.Logger.WithError(err).WithFields(logrus.Fields{
		// 							"date": i,
		// 							"type": x,
		// 						}).Error("failed to parse timestamp for type")
		// 						return
		// 					}

		// 					if date.Unix() >= start.Unix() && date.Unix() <= end.Unix() {

		// 						args = append(args, arg)
		// 						params = append(params, x, i, v)

		// 					}

		// 				}
		// 				if len(params) > 0 {
		// 					_, err = app.DB.Exec(fmt.Sprintf(`
		// 						INSERT IGNORE INTO prices (
		// 							type_id, date, price, created_at, updated_at
		// 						) VALUES %s
		// 					`, strings.Join(args, ",")), params...)
		// 					if err != nil {
		// 						app.Logger.WithError(err).WithFields(logrus.Fields{
		// 							"type_id": x,
		// 						}).Error("failed to insert record")
		// 						return
		// 					}
		// 				}

		// 				app.Logger.WithFields(logrus.Fields{
		// 					"type_id": x,
		// 				}).Info("successfully inserted historical record")
		// 				time.Sleep(time.Millisecond * 750)
		// 			})

		// 		}

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
	}
}

func main() {
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
