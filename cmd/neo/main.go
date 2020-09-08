package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	core "github.com/eveisesi/neo/app"
	"github.com/eveisesi/neo/server"
	"github.com/joho/godotenv"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/urfave/cli"
)

var (
	app *cli.App
)

func init() {
	_ = godotenv.Load(".env")

	app = cli.NewApp()
	app.Name = "neo"
	app.UsageText = "neo [parent] [child command] [--options]"
	app.Usage = "Service that manages all services related to Neo and its stable operation"
	app.Version = "v0.0.1"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name: "debug",
		},
	}
	app.Commands = []cli.Command{
		cli.Command{
			Name: "time",
			Action: func(c *cli.Context) error {
				fmt.Println(time.Now().Unix())
				return nil
			},
		},
		killmail(),
		alliances(),
		characters(),
		corporations(),
		cronCommand(),
		test(),
		cli.Command{
			Name:   "serve",
			Usage:  "Starts an HTTP Server to serve killmail data",
			Action: server.Action,
		},

		cli.Command{
			Name:  "listen",
			Usage: "Opens a WSS Connection to ZKillboard and lsitens to the stream",
			Action: func(c *cli.Context) error {
				_ = core.New("killmail-listener", false).Killmail.Websocket()

				return nil
			},
		},
		cli.Command{
			Name: "top",
			Action: func(c *cli.Context) error {
				return core.New("top", false).Top.Run()
			},
		},
		cli.Command{
			Name: "tracking",
			Action: func(c *cli.Context) error {

				app := core.New("tracking", false)

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

				app := core.New("autocompleter", false)
				txn := app.NewRelic.StartTransaction(app.Label)
				defer txn.End()
				ctx := newrelic.NewContext(context.Background(), txn)
				err := app.Search.Build(ctx)
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
				app := core.New("notifier", false)

				if !app.Config.SlackNotifierEnabled {
					return nil
				}

				app.Notification.Run(context.Background())
				return nil
			},
		},
		cli.Command{
			Name:        "updater",
			Description: "Updater ensures that all updatable records in the database are update date according to their CacheUntil timestamp.",
			Action: func(c *cli.Context) error {

				debug := c.GlobalBool("debug")

				app := core.New("updater", debug)

				var wg sync.WaitGroup

				wg.Add(1)
				go app.Character.UpdateExpired(context.Background())
				// wg.Add(1)
				// go app.Corporation.UpdateExpired(context.Background())
				// wg.Add(1)
				// go app.Alliance.UpdateExpired(context.Background())

				wg.Wait()

				return nil

			},
		},
		cli.Command{
			Name:        "market",
			Usage:       "Updates market prices in the Db",
			Subcommands: marketCommands(),
		},
	}
}

func main() {

	directories := []string{"./static/killmails/raw"}
	for _, directory := range directories {
		if _, err := os.Stat(directory); os.IsNotExist(err) {
			_ = os.MkdirAll(directory, 0666)
		}
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
