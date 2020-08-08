package main

import (
	"time"

	core "github.com/eveisesi/neo/app"
	"github.com/urfave/cli"
)

func migrateCommand() cli.Command {
	return cli.Command{
		Name: "migrate",
		Action: func(c *cli.Context) error {

			app := core.New("db-migrations", false)

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
	}
}
