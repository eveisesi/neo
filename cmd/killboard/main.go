package main

import (
	"log"
	"os"

	"github.com/eveisesi/neo/killmail/egress"
	"github.com/eveisesi/neo/killmail/ingress"
	"github.com/eveisesi/neo/server"
	"github.com/joho/godotenv"
	"github.com/urfave/cli"
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
	app.Name = "Killboard Core"
	app.Usage = "Service that manages all services related to Killboard and its stable operation"
	app.Version = "v0.0.1"
	app.Commands = []cli.Command{
		cli.Command{
			Name:   "ingress",
			Usage:  "Listen to a Redis PubSub channel for killmail hashes. On Message receive, reach out to CCP for Killmail Data and process.",
			Action: ingress.Action,
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
			Name:   "egress",
			Usage:  "Reaches out to the Zkillboard API and downloads historical killmail hashes, then reaches out to CCP for Killmail Data",
			Action: egress.Action,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "channel",
					Usage:    "channel is the key to use when  pulling killmail ids and hashes from redis to be resolved and inserted into the database",
					Required: true,
				},
				cli.StringFlag{
					Name:     "date",
					Usage:    "Date to use when request killmail hashes from zkillboard. (Format: YYYYMMDD)",
					Required: true,
				},
			},
		},
		cli.Command{
			Name:   "serve",
			Usage:  "Starts an HTTP Server to serve killmail data",
			Action: server.Action,
		},
	}
}

func main() {
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
