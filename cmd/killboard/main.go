package main

import (
	"log"
	"os"

	"github.com/ddouglas/killboard/killmail/egress"
	"github.com/ddouglas/killboard/killmail/ingress"
	"github.com/joho/godotenv"
	"github.com/urfave/cli"
)

var app *cli.App

func init() {
	err := godotenv.Load("cmd/killboard/.env")
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
					Usage:    "Channel to subscribe to using Redis Subscribe",
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
					Name:     "date",
					Usage:    "Date to use when request killmail hashes from zkillboard. (Format: YYYYMMDD)",
					Required: true,
				},
				cli.StringFlag{
					Name:     "channel",
					Usage:    "Channel to publish messages to using Redis Publish",
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
