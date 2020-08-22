package main

import (
	"context"
	"strconv"
	"time"

	"github.com/eveisesi/neo"
	core "github.com/eveisesi/neo/app"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/robfig/cron/v3"
	"github.com/urfave/cli"
)

func cronCommand() cli.Command {
	return cli.Command{
		Name:  "cron",
		Usage: "Spins up the crons",
		Action: func(ctx *cli.Context) error {
			app := core.New("cron", false)

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

			registerAutocompleteCron(c, app)
			registerEsiServerStatusCron(c, app)
			registerMarketDataCron(c, app)
			registerTrackingJanitorCron(c, app)

			c.Run()

			return nil
		},
	}
}

func registerAutocompleteCron(c *cron.Cron, app *core.App) {

	_, err := c.AddFunc("0 0 11 * * *", func() {
		txn := app.NewRelic.StartTransaction("cron-autocompleter")
		ctx := newrelic.NewContext(context.Background(), txn)
		app.Logger.WithContext(ctx).Info("rebuilding autocompleter index")

		err := app.Search.Build(ctx)
		if err != nil {
			app.Logger.WithContext(ctx).WithError(err).Error("failed to rebuild autocompleter index")
		}

		app.Logger.WithContext(ctx).Info("done rebuilding autocompleter index")
	})
	if err != nil {
		app.Logger.WithError(err).Fatal("failed to initialize autocompleteCron")
	}

}

func registerEsiServerStatusCron(c *cron.Cron, app *core.App) {

	_, err := c.AddFunc("*/30 * * * * *", func() {
		txn := app.NewRelic.StartTransaction("cron-esi-server-stats")
		ctx := newrelic.NewContext(context.Background(), txn)
		app.Logger.WithContext(ctx).Info("checking tq server status")

		serverStatus, m := app.ESI.GetStatus(ctx)
		if m.IsError() {
			app.Logger.WithContext(ctx).WithError(m.Msg).Error("Failed to fetch tq server status from ESI")
			return
		}

		if m.Code != 200 {
			app.Logger.WithContext(ctx).WithField("code", m.Code).Error("unable to acquire tq server status")
			return
		}

		app.Redis.WithContext(ctx).Set(neo.TQ_PLAYER_COUNT, serverStatus.Players, 0)
		app.Redis.WithContext(ctx).Set(neo.TQ_VIP_MODE, serverStatus.VIP.Bool, 0)

		app.Logger.WithContext(ctx).Info("done checking tq server status")
		txn.End()
	})
	if err != nil {
		app.Logger.WithError(err).Fatal("failed to initialize esiServerStatusCron")
	}

}

func registerMarketDataCron(c *cron.Cron, app *core.App) {

	_, err := c.AddFunc("0 10 11 * * *", func() {
		txn := app.NewRelic.StartTransaction("cron-market-data")
		ctx := newrelic.NewContext(context.Background(), txn)
		app.Logger.Info("starting fetch prices")
		app.Market.FetchPrices(ctx)
		app.Logger.Info("done with fetch prices")
		app.Logger.Info("starting fetch history ")
		app.Market.FetchHistory(ctx)
		app.Logger.Info("done with fetch history ")
		txn.End()
	})
	if err != nil {
		app.Logger.WithError(err).Fatal("failed to initialize marketDataCron")
	}

}

func registerTrackingJanitorCron(c *cron.Cron, app *core.App) {

	_, err := c.AddFunc("0 * * * * *", func() {
		txn := app.NewRelic.StartTransaction("cron-tracking-janitor")
		ctx := newrelic.NewContext(context.Background(), txn)
		app.Logger.WithContext(ctx).Info("starting esi tracking set janitor")

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
			a, err := app.Redis.WithContext(ctx).ZRemRangeByScore(set, "-inf", strconv.FormatInt(ts, 10)).Result()
			if err != nil {
				app.Logger.WithContext(ctx).WithError(err).Error("failed to fetch current count of esi success set from redis")
				return
			}
			count += a
		}

		app.Logger.WithContext(ctx).WithField("removed", count).Info("successfully cleared keys from success queue")
		app.Logger.WithContext(ctx).Info("stopping esi tracking set janitor")
		txn.End()
	})
	if err != nil {
		app.Logger.WithError(err).Fatal("failed to initialize trackingJanitorCron")
	}

}
