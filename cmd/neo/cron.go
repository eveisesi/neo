package main

import (
	"strconv"
	"time"

	"github.com/eveisesi/neo"
	core "github.com/eveisesi/neo/app"
	"github.com/robfig/cron/v3"
	"github.com/urfave/cli"
)

func cronCommand() cli.Command {
	return cli.Command{
		Name:  "cron",
		Usage: "Spins up the crons",
		Action: func(ctx *cli.Context) error {
			var err error
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

			_, err = c.AddFunc("0 10 11 * * *", marketDataCron)
			if err != nil {
				app.Logger.WithError(err).Fatal("failed to initialize marketDataCron")
			}

			_, err = c.AddFunc("*/30 * * * * *", esiServerStatusCron)
			if err != nil {
				app.Logger.WithError(err).Fatal("failed to initialize esiServerStatusCron")
			}

			_, err = c.AddFunc("0 * * * * *", trackingJanitorCron)
			if err != nil {
				app.Logger.WithError(err).Fatal("failed to initialize trackingJanitorCron")
			}

			_, err = c.AddFunc("0 0 11 * * *", autocompleteCron)
			if err != nil {
				app.Logger.WithError(err).Fatal("failed to initialize autocompleteCron")
			}

			c.Run()

			return nil
		},
	}
}

func autocompleteCron() {
	app := core.New()

	app.Logger.Info("rebuilding autocompleter index")

	err := app.Search.Build()
	if err != nil {
		app.Logger.WithError(err).Error("failed to rebuild autocompleter index")
	}

	app.Logger.Info("done rebuilding autocompleter index")
}

func esiServerStatusCron() {
	app := core.New()

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

}

func marketDataCron() {
	app := core.New()

	app.Logger.Info("starting fetch prices")
	app.Market.FetchPrices()
	app.Logger.Info("done with fetch prices")
	app.Logger.Info("starting fetch history ")
	app.Market.FetchHistory()
	app.Logger.Info("done with fetch history ")
}

func trackingJanitorCron() {
	app := core.New()

	app.Logger.Info("starting esi tracking set janitor")

	ts := time.Now().Add(time.Minute * -6).UnixNano()
	sets := []string{
		neo.REDIS_ESI_TRACKING_OK,
		neo.REDIS_ESI_TRACKING_NOT_MODIFIED,
		neo.REDIS_ESI_TRACKING_CALM_DOWN,
		neo.REDIS_ESI_TRACKING_4XX,
		neo.REDIS_ESI_TRACKING_5XX,
	}
	count := int64(0)
	for _, set := range sets {
		a, err := app.Redis.ZRemRangeByScore(set, "-inf", strconv.FormatInt(ts, 10)).Result()
		if err != nil {
			app.Logger.WithError(err).Error("failed to fetch current count of esi success set from redis")
			return
		}
		count += a
	}

	app.Logger.WithField("removed", count).Info("successfully cleared keys from success queue")
	app.Logger.Info("stopping esi tracking set janitor")

}
